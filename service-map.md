# Service Map

Map nhanh khu vực nghiệp vụ → đường dẫn code. Mục đích: AI/người đọc tìm code liên quan mà không cần grep toàn repo. Cập nhật mỗi khi thêm/sửa/xoá file có ý nghĩa nghiệp vụ (theo quy ước trong [CLAUDE.md](CLAUDE.md)).

Khung bên dưới theo đúng layout đã chốt trong system-architecture.md — điền dần khi code thật xuất hiện. Trạng thái hiện tại: **order-service** đã xong CRUD order + publish Kafka; inventory-service, notification-service, frontend chưa có code.

---

## order-service

- Entry point, wiring router/DB: [order-service/main.go](order-service/main.go)
- Config (đọc `PORT`, `DATABASE_URL`, `KAFKA_BROKERS`, `KAFKA_ORDER_CREATED_TOPIC` từ env): [order-service/internal/config/config.go](order-service/internal/config/config.go)
- Kết nối Postgres + tạo schema `orders`/`order_items`: [order-service/internal/db/db.go](order-service/internal/db/db.go)
- Model `Order`, `OrderItem`, request DTO, `OrderCreatedEvent` (payload publish Kafka): [order-service/internal/models/order.go](order-service/internal/models/order.go)
- Kafka producer, publish `order-created` (fire-and-forget, key = `order_id`): [order-service/internal/kafka/producer.go](order-service/internal/kafka/producer.go)
- Tạo order → `POST /api/orders`, publish event sau khi commit transaction: [order-service/internal/handlers/order.go](order-service/internal/handlers/order.go) (`CreateOrder`)
- Lấy 1 order → `GET /api/orders/:id`: [order-service/internal/handlers/order.go](order-service/internal/handlers/order.go) (`GetOrder`)
- Danh sách order theo customer → `GET /api/orders?customer_id=`: [order-service/internal/handlers/order.go](order-service/internal/handlers/order.go) (`ListOrders`)

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

- `infra/docker-compose.yml`: chạy Postgres (`order_db`), Kafka KRaft mode — không Zookeeper (Confluent image, 1 node broker+controller, topic `order-created` 3 partitions) cho local dev — sẽ bổ sung Redis, các service khác, monitoring stack ở các bước sau.
- `infra/k8s/`: *(chưa có code)*
- `infra/grafana/`: *(chưa có code)*
