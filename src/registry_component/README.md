## Компонент реестров

Логический компонент для операций:

- `register_operator`
- `register_customer`

Реализация пока остаётся in-process и вызывается через gateway, а маршрутизация
определена в `internal/gateway/gateway.go`.
