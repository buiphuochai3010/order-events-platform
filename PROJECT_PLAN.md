# Order System — Microservice + Kafka + K8s + Monitoring

Project cá nhân để thực hành: microservice architecture, event-driven với Kafka, deploy K8s local, và monitoring stack (Prometheus + Grafana). Mục tiêu: repo tự chạy được từ A-Z trên máy local, không tốn tiền cloud, và có dashboard để mở lên sau vài tháng vẫn biết ngay hệ thống có vấn đề gì không.

---

## 1. Kiến trúc tổng quan

```
Frontend (React + Tailwind + Ant Design)
      │
      ▼
Order Service (Gin) ──publish──► Kafka Topic: "order-created"
      │                                    │
      ▼ (ghi DB)                ┌──────────┴──────────┐
   PostgreSQL                   ▼                     ▼
                          Inventory Service      Notification Service
                          (Gin, consumer)         (Gin, consumer)
                               │                        │
                               ▼                        ▼
                          PostgreSQL/Redis        Log ra / giả lập gửi email
```

**Vì sao 3 service, không hơn:** đủ để thể hiện 1 event chảy qua nhiều consumer độc lập — đúng bản chất microservice. Thêm service chỉ tốn thời gian, không tăng thêm signal khi phỏng vấn.

---

## 2. Vai trò từng thành phần

### Kafka — trục truyền tin giữa các service
- Order Service chỉ cần **publish** message vào topic `order-created`, không cần biết ai đang lắng nghe, không cần chờ phản hồi.
- Inventory Service và Notification Service tự **subscribe (consume)** topic đó, xử lý độc lập theo tốc độ riêng.
- Lợi ích để kể trong interview: decouple service, tăng khả năng chịu lỗi — nếu Notification Service chết, Order vẫn tạo được đơn bình thường, message vẫn nằm trong Kafka chờ xử lý khi Notification sống lại.

### K8s — quản lý vận hành từng service
- Mỗi service (Order, Inventory, Notification, Kafka, Postgres) là 1 **Pod**, quản lý qua **Deployment** (tự restart nếu pod chết, scale replica).
- **Service** (K8s object) để các pod gọi nhau qua tên nội bộ (`order-service.default.svc.cluster.local`) thay vì IP cứng.
- **ConfigMap/Secret** quản lý biến môi trường (DB connection string, Kafka broker address) tách khỏi code.
- Thực hành: viết `deployment.yaml`, `service.yaml` cho mỗi service, `kubectl apply -f .`, test tự hồi sinh bằng `kubectl delete pod xxx`.

### Prometheus + Grafana — monitoring
- Mỗi service Gin expose endpoint `/metrics` (dùng `prometheus/client_golang`): request/sec, latency, lỗi 5xx, **Kafka consumer lag** (đo Inventory/Notification xử lý kịp message hay bị tồn đọng).
- **Prometheus** scrape `/metrics` mỗi 15s, lưu theo thời gian.
- **Grafana** đọc Prometheus, vẽ dashboard tổng quan: latency từng service, tổng request, Kafka lag.
- Mục tiêu: mở Grafana lên sau 2 tháng, thấy ngay ví dụ "Inventory Service lag Kafka tăng vọt" → biết service đó đang nghẽn hoặc down.

---

## 3. Chạy ở đâu — tránh tốn tiền

- Kafka và K8s là software mã nguồn mở, free. Tốn tiền chỉ khi thuê managed service (GKE, Confluent Cloud...).
- **Kafka local**: `docker-compose` (Kafka + Zookeeper, hoặc Redpanda — nhẹ hơn, tương thích Kafka API).
- **K8s local**: **k3d** hoặc **Minikube** — giả lập cluster trong Docker, `kubectl apply` y như cluster thật.
- Đẩy lên GitHub là code + config (Dockerfile, docker-compose.yml, manifest .yaml, CI/CD workflow) — người xem tự chạy lại bằng `docker-compose up` hoặc `k3d cluster create`, không cần trả tiền hosting để demo.
- Nếu muốn có link demo online 24/7: deploy bản đơn giản (không Kafka, không K8s) lên free tier Railway/Render/Fly.io, kèm video demo ngắn cho bản đầy đủ.

---

## 4. Stack cụ thể

| Layer | Công nghệ |
|---|---|
| Backend | Golang + Gin |
| Database | PostgreSQL |
| Message broker | Kafka (hoặc Redpanda) |
| Container orchestration | K8s (k3d/Minikube local) |
| Monitoring | Prometheus + Grafana |
| Log tập trung | Loki + Promtail |
| Alert | Alertmanager → Telegram/Discord webhook |
| CI/CD | GitHub Actions |
| Frontend | ReactJS + Tailwind + Ant Design (vibe-code, không đầu tư nhiều) |

---

## 5. CI/CD pipeline (GitHub Actions)

```
push → lint → unit test (go test -cover, threshold ~70%) → build docker image → push registry → deploy (kubectl apply / helm upgrade)
```

- Tách riêng stage: test trên PR, deploy trên merge vào main — thể hiện hiểu branch protection/environment tách biệt.

---

## 6. Thứ tự build (để không bị ngộp)

1. **Order Service** (Gin + Postgres) — chạy được, tạo order, chưa có Kafka.
2. Thêm **Kafka** (docker-compose), Order Service publish event khi tạo order.
3. Viết **Inventory Service** + **Notification Service**, consume message.
4. Chạy ổn toàn bộ trên **docker-compose** trước (chưa cần K8s).
5. Chuyển sang **K8s local (k3d)** — viết manifest, deploy lại đúng 3 service.
6. Thêm **Prometheus + Grafana + Loki + Alertmanager** cuối cùng — lúc này có sẵn traffic thật để vẽ số liệu.

---

## 7. Phân bổ effort ưu tiên

- Backend (Gin + Postgres + Kafka): **40%**
- Monitoring stack: **30%**
- CI/CD: **20%**
- K8s deploy (chỉ cần chạy được, không cần đẹp): **10%**
- Frontend vibe-code: thời gian còn lại

---

## 8. Checklist README cần có khi xong

- [ ] Cách chạy local (`docker-compose up -d`, `k3d cluster create`)
- [ ] Diagram kiến trúc (bản trên)
- [ ] Giải thích tại sao dùng Kafka ở đây (event-driven, decouple)
- [ ] Screenshot/video ngắn dashboard Grafana
- [ ] Link CI/CD pipeline (GitHub Actions badge)
- [ ] Test coverage badge
