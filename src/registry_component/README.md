## Компонент реестров

Компонент обрабатывает операции:

- `register_operator`
- `register_customer`

Код обработки входящих bus-сообщений находится в `component.go`.
Вызов компонента идёт через gateway и общий bus-диспетчер в
`internal/bus/handler`.
