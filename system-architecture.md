# System Architecture — order-events-platform

Tài liệu này chi tiết hoá [PROJECT_PLAN.md](PROJECT_PLAN.md) thành thiết kế kỹ thuật cụ thể: repo layout, data model, event schema, API contract, và cách các thành phần giao tiếp. Đây là thiết kế ban đầu — sẽ cập nhật khi build thực tế phát sinh thay đổi.

---

## 1. Repo layout (monorepo)

```
order-events-platform/
├── order-service/           # Gin, publish event
├── inventory-service/       # Gin, consumer, Postgres + Redis cache
├── notification-service/    # Gin, consumer, log/giả lập gửi email
├── frontend/                 # Next.js + Tailwind + Ant Design
├── infra/
│   ├── docker-compose.yml    # Kafka/Redpanda, Postgres, Redis, 3 service, Prometheus, Grafana, Loki
│   ├── k8s/
│   │   ├── order-service/    # deployment.yaml, service.yaml, configmap.yaml
│   │   ├── inventory-service/
│   │   ├── notification-service/
│   │   └── monitoring/       # prometheus, grafana, loki, alertmanager manifests
│   └── grafana/               # dashboard JSON provisioning
├── .github/workflows/         # CI/CD
├── PROJECT_PLAN.md
└── system-architecture.md
```

Mỗi service là 1 Go module độc lập (`go.mod` riêng) để build Docker image riêng và có thể tách thành repo riêng sau này nếu cần.

---

## 2. Data model

### 2.1 Database strategy

Dùng **1 Postgres instance, tách DB theo service** (không phải database-per-service instance riêng) — đủ để thể hiện ranh giới dữ liệu giữa các service, nhẹ tài nguyên khi chạy local.

```
Postgres instance (1 container)
├── order_db          (Order Service)
└── inventory_db       (Inventory Service)
```

Mỗi service chỉ có connection string tới DB của mình (qua ConfigMap/Secret) — **không service nào được query trực tiếp DB của service khác**. Đây là ranh giới quan trọng nhất để giữ đúng tinh thần microservice, dù chạy chung 1 Postgres.

### 2.2 Order Service — `order_db`

```sql
CREATE TABLE orders (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL,
    status      VARCHAR(20) NOT NULL DEFAULT 'created', -- created | confirmed | cancelled
    total_amount NUMERIC(12,2) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE order_items (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id    UUID NOT NULL REFERENCES orders(id),
    product_id  UUID NOT NULL,
    quantity    INT NOT NULL,
    unit_price  NUMERIC(12,2) NOT NULL
);
```

### 2.3 Inventory Service — `inventory_db` (Postgres) + Redis cache

- **Postgres** là nguồn sự thật (source of truth) cho tồn kho.
- **Redis** cache số lượng tồn (`stock:{product_id}`) để đọc nhanh; ghi/trừ tồn kho vẫn qua Postgres trước, sau đó cập nhật lại cache (cache-aside, không cache-through).

```sql
CREATE TABLE stock (
    product_id UUID PRIMARY KEY,
    quantity   INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### 2.4 Notification Service

Không cần DB — nhận event, log ra console (hoặc ghi file log) giả lập gửi email. Không lưu trạng thái.

---

## 3. Kafka — event schema

### Topic: `order-created`

- **Key**: `order_id` (đảm bảo message cùng 1 order vào cùng partition, giữ thứ tự)
- **Partitions**: 3 (đủ để demo song song, không cần nhiều hơn cho local)
- **Value** (JSON):

```json
{
  "event_id": "uuid",
  "event_type": "order.created",
  "occurred_at": "2026-07-05T10:00:00Z",
  "order_id": "uuid",
  "customer_id": "uuid",
  "items": [
    { "product_id": "uuid", "quantity": 2, "unit_price": 100000 }
  ],
  "total_amount": 200000
}
```

`event_id` dùng để consumer chống xử lý trùng (idempotency) nếu Kafka redeliver message — nhưng ở mức tối giản của project này, **không bắt buộc consumer phải idempotent ngay từ đầu**; đây là điểm có thể nêu ra khi phỏng vấn như "biết vấn đề, chưa cần giải ở giai đoạn demo".

### Consumer groups

- `inventory-service-group` — trừ tồn kho theo `items`.
- `notification-service-group` — log giả lập gửi email xác nhận đơn.

Hai group độc lập, cùng đọc topic `order-created`, không ảnh hưởng lẫn nhau.

### Error handling (tối giản, theo quyết định giai đoạn này)

Không dùng retry/DLQ. Nếu consumer xử lý message lỗi (vd DB down), **log lỗi và bỏ qua**, tiếp tục message tiếp theo. Đây là giới hạn có chủ đích của giai đoạn đầu — ghi rõ trong README như một trade-off đã biết, không phải thiếu sót. Nếu sau này muốn nâng cấp: thêm retry với backoff + topic `order-created.dlq`.

---

## 4. API contract — Order Service (REST, expose cho Frontend)

| Method | Path | Mô tả |
|---|---|---|
| POST | `/api/orders` | Tạo order mới → ghi `order_db` → publish `order-created` |
| GET | `/api/orders/:id` | Lấy chi tiết 1 order |
| GET | `/api/orders?customer_id=` | Danh sách order theo customer |
| GET | `/metrics` | Prometheus scrape endpoint |
| GET | `/healthz` | Liveness/readiness probe cho K8s |

Inventory Service và Notification Service **không expose API cho Frontend** — chỉ có `/metrics` và `/healthz` (nội bộ, phục vụ K8s probe + Prometheus).

**Luồng tạo order** (đồng bộ với client, publish Kafka là fire-and-forget):
1. Frontend `POST /api/orders`.
2. Order Service validate → ghi `orders` + `order_items` vào `order_db` (transaction).
3. Order Service publish event `order-created` lên Kafka (không chờ consumer xử lý).
4. Trả `201 Created` cho Frontend ngay sau bước 2–3, **không đợi** Inventory/Notification xử lý xong — đúng bản chất async/decouple.

---

## 5. Hạ tầng & topology

### 5.1 Docker Compose (local dev)

```
docker-compose.yml
├── postgres          (order_db + inventory_db, 2 DB trong 1 container)
├── redis
├── kafka + zookeeper (hoặc redpanda)
├── order-service
├── inventory-service
├── notification-service
├── prometheus
├── grafana
├── loki + promtail
└── alertmanager
```

Network nội bộ: các service gọi nhau qua tên container (`order-service:8080`, `postgres:5432`, `kafka:9092`).

### 5.2 K8s (k3d/Minikube local)

Mỗi service có `deployment.yaml` + `service.yaml` riêng trong `infra/k8s/<service>/`. Postgres, Redis, Kafka chạy dạng Deployment + PersistentVolumeClaim đơn giản (không cần StatefulSet phức tạp cho mục tiêu demo). ConfigMap chứa non-secret config (Kafka broker address, DB host), Secret chứa DB password.

Service discovery nội bộ qua DNS chuẩn K8s: `order-service.default.svc.cluster.local`.

### 5.3 Monitoring

- Mỗi Go service expose `/metrics` qua `prometheus/client_golang`: `http_requests_total`, `http_request_duration_seconds`, và riêng Inventory/Notification thêm `kafka_consumer_lag`.
- Prometheus scrape 15s/lần.
- Grafana dashboard: latency theo service, tổng request, Kafka consumer lag theo group.
- Loki + Promtail gom log tập trung (log dạng JSON structured từ Gin).
- Alertmanager bắn cảnh báo qua Telegram/Discord webhook khi: lag Kafka vượt ngưỡng, tỷ lệ lỗi 5xx cao, hoặc service down.

---

## 6. CI/CD

```
push → lint (golangci-lint) → go test -cover (ngưỡng ~70%) → build Docker image (mỗi service) → push registry → deploy
```

- PR: chạy lint + test, không deploy.
- Merge vào `main`: build + push image + `kubectl apply` (hoặc `helm upgrade` nếu sau này đóng gói Helm chart).
- Matrix build theo từng service (3 job song song) vì mỗi service là Go module riêng.

---

## 7. Giới hạn thiết kế đã biết (ghi rõ để không bị hỏi bất ngờ khi phỏng vấn)

- Không có distributed transaction / saga giữa Order và Inventory — nếu Inventory trừ kho fail, Order vẫn ở trạng thái `created`, không tự động rollback. Đây là trade-off chấp nhận được cho phạm vi demo.
- Consumer không idempotent — message trùng (do Kafka redeliver) có thể gây trừ kho 2 lần. Biết vấn đề, chưa xử lý ở bản đầu.
- Không có API Gateway / auth ở Order Service — Frontend gọi thẳng. Nếu mở rộng, đây là điểm thêm vào sau (JWT, rate limit).
