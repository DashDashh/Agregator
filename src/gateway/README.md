## Gateway

Gateway слушает системный топик `systems.agregator`, маршрутизирует входящие
сообщения по полю `action` во внутренние компоненты и публикует ответы в
`components.agregator.responses`.

Текущая таблица маршрутизации находится в `src/gateway/bus/gateway`, а реальные
обработчики сообщений находятся в пакетах внутри `src/`.

Gateway связывает действия со следующими компонентами:

- `registry_component`
- `orders_component`
- `contracts_component`
- `analytics_component`

HTTP routing, middleware, publisher fan-out and configuration also live inside
`src/gateway`, because this is the system entry adapter. Business HTTP handlers
live in their owning components:

- `src/registry_component/httpapi`
- `src/orders_component/httpapi`
- `src/contracts_component/httpapi`
