## Компонент реестров

Компонент обрабатывает операции:

- `register_operator`
- `register_customer`

Код обработки входящих bus-сообщений находится в `component.go`.
Вызов компонента идёт через gateway и общий bus-диспетчер в
`src/gateway/bus/handler`.

Код auth находится в `src/registry_component/auth`, потому что регистрация,
идентификация и роли относятся к домену Registry.

HTTP-обработчики регистрации, профилей и входа находятся в
`src/registry_component/httpapi`.
