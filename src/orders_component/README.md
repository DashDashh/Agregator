## Компонент заказов

Компонент обрабатывает операции:

- `create_order`
- `select_executor`
- `auto_search_executor`
- сценариев работы с `price_offer` и `confirm_price`

Код обработки входящих bus-сообщений находится в `component.go`.
HTTP-обработчики заказов находятся в `src/orders_component/httpapi`, а persistence
адаптер временно вынесен в `src/shared/store`, чтобы сохранить текущее поведение
до полного выделения компонента в отдельный сервис.
