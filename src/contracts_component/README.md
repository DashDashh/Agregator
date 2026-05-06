## Компонент контрактов

Компонент обрабатывает операции:

- `conclude_contract`
- `confirm_execution`
- `create_dispute`

Код обработки входящих bus-сообщений находится в `component.go`.
Компонент вызывается через gateway и общий bus-диспетчер в
`src/gateway/bus/handler`.

HTTP-обработчики подтверждения цены, предложения цены и подтверждения выполнения
находятся в `src/contracts_component/httpapi`.
