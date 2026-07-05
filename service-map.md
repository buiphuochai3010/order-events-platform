# Service Map

Map nhanh khu vực nghiệp vụ → đường dẫn code. Mục đích: AI/người đọc tìm code liên quan mà không cần grep toàn repo. Cập nhật mỗi khi thêm/sửa/xoá file có ý nghĩa nghiệp vụ (theo quy ước trong [CLAUDE.md](CLAUDE.md)).

Trạng thái hiện tại: **chưa có code**, repo mới có docs (PROJECT_PLAN.md, system-architecture.md). Khung bên dưới theo đúng layout đã chốt trong system-architecture.md — điền dần khi code thật xuất hiện.

---

## order-service

- Entry point, wiring router/DB: [order-service/main.go](order-service/main.go)
- Config (đọc `PORT`, `DATABASE_URL` từ env): [order-service/internal/config/config.go](order-service/internal/config/config.go)
- Kết nối Postgres + tạo schema `orders`/`order_items`: [order-service/internal/db/db.go](order-service/internal/db/db.go)
- Model `Order`, `OrderItem`, request DTO: [order-service/internal/models/order.go](order-service/internal/models/order.go)
- Tạo order → `POST /api/orders`: [order-service/internal/handlers/order.go](order-service/internal/handlers/order.go) (`CreateOrder`)
- Lấy 1 order → `GET /api/orders/:id`: [order-service/internal/handlers/order.go](order-service/internal/handlers/order.go) (`GetOrder`)
- Danh sách order theo customer → `GET /api/orders?customer_id=`: [order-service/internal/handlers/order.go](order-service/internal/handlers/order.go) (`ListOrders`)
- Publish event `order-created`: *(chưa làm — theo thứ tự PROJECT_PLAN.md, sẽ thêm ở bước Kafka)*

## inventory-service
*(chưa có code)*

- Consumer group `inventory-service-group`, xử lý `order-created`:
- Trừ tồn kho (Postgres) + cập nhật cache (Redis):
- Model `stock`:

## notification-service
*(chưa có code)*

- Consumer group `notification-service-group`, xử lý `order-created`:
- Giả lập gửi email (log):

## frontend
*(chưa có code)*

## infra

- `infra/docker-compose.yml`: chạy Postgres (`order_db`) cho local dev — sẽ bổ sung Kafka, Redis, các service khác, monitoring stack ở các bước sau.
- `infra/k8s/`: *(chưa có code)*
- `infra/grafana/`: *(chưa có code)*
