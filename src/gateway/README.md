## Gateway

Gateway слушает системный топик `systems.agregator`, маршрутизирует входящие
сообщения по полю `action` во внутренние компоненты и публикует ответы в
`components.agregator.responses`.

Текущая таблица маршрутизации находится в
`internal/gateway/gateway.go` и связывает действия со следующими логическими
компонентами:

- `registry_component`
- `orders_component`
- `contracts_component`
- `analytics_component`
