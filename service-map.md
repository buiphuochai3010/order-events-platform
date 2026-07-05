# Service Map

Map nhanh khu vực nghiệp vụ → đường dẫn code. Mục đích: AI/người đọc tìm code liên quan mà không cần grep toàn repo. Cập nhật mỗi khi thêm/sửa/xoá file có ý nghĩa nghiệp vụ (theo quy ước trong [CLAUDE.md](CLAUDE.md)).

Trạng thái hiện tại: **chưa có code**, repo mới có docs (PROJECT_PLAN.md, system-architecture.md). Khung bên dưới theo đúng layout đã chốt trong system-architecture.md — điền dần khi code thật xuất hiện.

---

## order-service
*(chưa có code)*

- Tạo order → `POST /api/orders`:
- Lấy order → `GET /api/orders/:id`:
- Publish event `order-created`:
- Model `orders`, `order_items`:

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
*(chưa có code)*

- `infra/docker-compose.yml`:
- `infra/k8s/`:
- `infra/grafana/`:
