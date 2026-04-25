## Компонент контрактов

Компонент обрабатывает операции:

- `conclude_contract`
- `confirm_execution`
- `create_dispute`

Код обработки входящих bus-сообщений находится в `component.go`.
Компонент вызывается через gateway и общий bus-диспетчер в
`internal/bus/handler`.
