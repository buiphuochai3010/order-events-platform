# CLAUDE.md

Hướng dẫn cho Claude Code khi làm việc trong repo `order-events-platform`.

## Tài liệu tham khảo
- [PROJECT_PLAN.md](PROJECT_PLAN.md) — mục tiêu, stack, thứ tự build, phân bổ effort.
- [system-architecture.md](system-architecture.md) — thiết kế chi tiết: data model, event schema Kafka, API contract, infra topology.
- [service-map.md](service-map.md) — map nhanh: khu vực nghiệp vụ → đường dẫn code.

## Quy ước khi sửa code
- Thêm/xoá/di chuyển service, đổi event schema, đổi DB schema → cập nhật [system-architecture.md](system-architecture.md) trước hoặc cùng lúc.
- Thêm/sửa file có ý nghĩa nghiệp vụ (handler, model, consumer, publisher...) → cập nhật mục tương ứng trong [service-map.md](service-map.md).
- Không tự thêm service/thư viện/abstraction ngoài phạm vi PROJECT_PLAN.md nếu chưa hỏi — đây là project học tập có phạm vi cố định (3 service), không phải sản phẩm cần mở rộng liên tục.
