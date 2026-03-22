# Agregator Insurer

Агрегатор страховых заявок на доставку дронами.

## Стек

- **Go 1.24** — сервис
- **PostgreSQL 16** — хранение заказчиков, эксплуатантов и заказов
- **Apache Kafka** — основной транспорт сообщений
- **MQTT (Mosquitto)** — дополнительный транспорт для обмена с эксплуатантами
- **Docker Compose** — запуск всего окружения одной командой

## Запуск

```bash
docker compose up -d --build
```

Сервис поднимется на `http://localhost:8080`.

## Режим транспорта

Для обмена с эксплуатантами (`operator.requests` / `operator.responses`) режим выбирается через переменную окружения `OPERATOR_TRANSPORT`:

- `kafka` — только Kafka (режим по умолчанию)
- `both` — Kafka + MQTT одновременно

> Важно: контур `aggregator.requests` / `aggregator.responses` по-прежнему работает через Kafka. Поэтому на текущей архитектуре поддерживаются именно режимы `kafka` и `both`, а не `mqtt only`.

Пример для `docker-compose.yml`:

```yaml
environment:
  OPERATOR_TRANSPORT: ${OPERATOR_TRANSPORT:-kafka}
```

Порядок старта автоматический: сначала PostgreSQL (с healthcheck), затем Kafka, затем агрегатор. MQTT-брокер поднимается как отдельный сервис и используется агрегатором только при выборе режима `both`.

> Тогда можно запускать `export OPERATOR_TRANSPORT=both && docker compose up -d --build`.


## API

### Проверка здоровья

```
GET /health
```

**Ответ:**

```json
{"status": "ok"}
```

---

### Заказчики

#### Зарегистрировать заказчика

```
POST /customers
```

**Тело запроса:**

```json
{
  "name": "Иван Иванов",
  "email": "ivan@mail.ru",
  "phone": "+79001234567"
}
```

**Ответ `201`:**

```json
{
  "id": "b9e8b4d6-2318-429b-944c-11e46db1fbfe",
  "name": "Иван Иванов",
  "email": "ivan@mail.ru",
  "phone": "+79001234567"
}
```

---

### Эксплуатанты

#### Зарегистрировать эксплуатанта

```
POST /operators
```

**Тело запроса:**

```json
{
  "name": "ООО АэроДоставка",
  "license": "LIC-2024-001",
  "email": "ops@aerodostavka.ru"
}
```

**Ответ `201`:**

```json
{
  "id": "a1b2c3d4-...",
  "name": "ООО АэроДоставка",
  "license": "LIC-2024-001",
  "email": "ops@aerodostavka.ru"
}
```

---

### Заказы

#### Создать заказ

```
POST /orders
```

> Требует существующего `customer_id`. При создании заказ автоматически отправляется эксплуатантам через выбранный транспорт (`operator.requests`): Kafka либо Kafka+MQTT.

**Тело запроса (delivery — по умолчанию):**

```json
{
  "customer_id": "b9e8b4d6-2318-429b-944c-11e46db1fbfe",
  "description": "Доставить документы из офиса на склад",
  "budget": 2500.00,
  "mission_type": "delivery",
  "security_goals": ["ЦБ1", "ЦБ3"],
  "from_lat": 55.7558,
  "from_lon": 37.6173,
  "to_lat": 55.8000,
  "to_lon": 37.6500
}
```

**Тело запроса (agro):**

```json
{
  "customer_id": "b9e8b4d6-2318-429b-944c-11e46db1fbfe",
  "description": "Обработать поле",
  "budget": 4000.00,
  "mission_type": "agro",
  "security_goals": ["ЦБ2", "ЦБ4"],
  "top_left_lat": 55.90,
  "top_left_lon": 37.40,
  "bottom_right_lat": 55.80,
  "bottom_right_lon": 37.60
}
```

**Ответ `201`:**

```json
{
  "id": "e16d6d12-b045-4eb9-bf07-b811a3836e57",
  "customer_id": "b9e8b4d6-2318-429b-944c-11e46db1fbfe",
  "description": "Доставить документы из офиса на склад",
  "budget": 2500,
  "from_lat": 55.7558,
  "from_lon": 37.6173,
  "to_lat": 55.8,
  "to_lon": 37.65,
  "mission_type": "delivery",
  "security_goals": ["ЦБ1", "ЦБ3"],
  "top_left_lat": 0,
  "top_left_lon": 0,
  "bottom_right_lat": 0,
  "bottom_right_lon": 0,
  "commission_amount": 0,
  "operator_amount": 0,
  "status": "pending",
  "created_at": "2026-03-04T17:31:12.658581072Z"
}
```

---

#### Получить список всех заказов

```
GET /orders
```

**Ответ `200`:** массив объектов заказа (sorted by `created_at DESC`).

---

#### Получить заказ по ID

```
GET /orders/{id}
```

**Ответ `200`:** объект заказа.

**Ответ `404`:**

```json
{"error": "заказ не найден"}
```

---

#### Подтвердить цену эксплуатанта

```
POST /orders/{id}/confirm-price
```

> Пользователь принимает оферту от эксплуатанта. Агрегатор переводит заказ в статус `confirmed`
> и отправляет эксплуатанту сообщение `confirm_price` через выбранный транспорт (`operator.requests`).

**Тело запроса:**

```json
{
  "operator_id": "a1b2c3d4-...",
  "accepted_price": 2200.00
}
```

**Ответ `200` (учитывает сервисный сбор):**

```json
{
  "order_id": "e16d6d12-...",
  "operator_id": "a1b2c3d4-...",
  "accepted_price": 2200.00,
  "commission_amount": 220.0,
  "operator_amount": 1980.0,
  "status": "confirmed"
}
```

> Сбор считается как `accepted_price * COMMISSION_RATE` (env, по умолчанию 0.1). Оператор получает `accepted_price - commission_amount`.

#### Подтвердить выполнение заказчиком

```
POST /orders/{id}/confirm-completion
```

**Ответ `200`:**

```json
{
  "order_id": "...",
  "status": "completed"
}
```

---

**Статусы заказа:**

| Статус                       | Когда выставляется                                                                                                       |
| ---------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| `pending`                        | Заказ создан, ждёт предложений                                                                                  |
| `searching`                      | Агрегатор опубликовал заказ в `operator.requests` через выбранный транспорт            |
| `matched`                        | Эксплуатант прислал оферту цены (`price_offer`)                                                             |
| `confirmed`                      | Пользователь принял цену (`POST .../confirm-price`)                                                               |
| `completed_pending_confirmation` | Оператор сообщил об успехе (`order_result` success=true), ждём подтверждения заказчика |
| `completed`                      | Заказчик подтвердил выполнение (`POST .../confirm-completion`)                                              |
| `dispute`                        | Эксплуатант сообщил о срыве (`order_result` success=false)                                                      |

---

## Пример полного запроса

```bash
# 1. Создать заказчика
CUSTOMER_ID=$(curl -s -X POST http://localhost:8080/customers \
  -H "Content-Type: application/json" \
  -d '{"name":"Иван Иванов","email":"ivan@example.com","phone":"+79001234567"}' \
  | jq -r .id)
echo "CUSTOMER_ID=$CUSTOMER_ID"

# 2. Создать эксплуатанта
OPERATOR_ID=$(curl -s -X POST http://localhost:8080/operators \
  -H "Content-Type: application/json" \
  -d '{"name":"ООО Дроны","license":"LIC-001","email":"ops@example.com"}' \
  | jq -r .id)
echo "OPERATOR_ID=$OPERATOR_ID"

# 3. Создать заказ (delivery) — уйдёт в operator.requests через выбранный транспорт
ORDER_ID=$(curl -s -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id":"'"'"$CUSTOMER_ID'"'"",
    "description":"Доставить документы из офиса на склад",
    "budget":3000,
    "mission_type":"delivery",
    "security_goals":["ЦБ1"],
    "from_lat":55.7558,"from_lon":37.6173,
    "to_lat":55.8000,"to_lon":37.6500,
    "top_left_lat":0,"top_left_lon":0,
    "bottom_right_lat":0,"bottom_right_lon":0
  }' \
  | jq -r .id)
echo "ORDER_ID=$ORDER_ID"

# 4. Подтвердить цену эксплуатанта (учитывается COMMISSION_RATE)
curl -s -X POST http://localhost:8080/orders/$ORDER_ID/confirm-price \
  -H "Content-Type: application/json" \
  -d '{"operator_id":"'"'"$OPERATOR_ID'"'"","accepted_price":4500}' | jq

# 5. Подтвердить выполнение заказчиком
curl -s -X POST http://localhost:8080/orders/$ORDER_ID/confirm-completion \
  -H "Content-Type: application/json" -d '{}' | jq

# 6. Проверить заказ и список
curl -s http://localhost:8080/orders/$ORDER_ID | jq
curl -s http://localhost:8080/orders | jq
```

---

## Форматы сообщений

### Агрегатор → Эксплуатант (`<prefix>.operator.requests`)

Все сообщения завёрнуты в стандартный конверт:

```json
{
  "request_id": "<order_id>",
  "type": "<тип>",
  "payload": { ... }
}
```

#### `create_order` — новый заказ (отправляется при `POST /orders`)

```json
{
  "request_id": "e16d6d12-b045-4eb9-bf07-b811a3836e57",
  "type": "create_order",
  "payload": {
    "customer_id": "b9e8b4d6-2318-429b-944c-11e46db1fbfe",
    "description": "Доставить документы из офиса на склад",
    "budget": 3000.00,
    "mission_type": "delivery",
    "security_goals": ["ЦБ1"],
    "from_lat": 55.7558,
    "from_lon": 37.6173,
    "to_lat": 55.8000,
    "to_lon": 37.6500,
    "top_left_lat": 0,
    "top_left_lon": 0,
    "bottom_right_lat": 0,
    "bottom_right_lon": 0
  }
}
```

#### `confirm_price` — пользователь принял цену (отправляется при `POST /orders/{id}/confirm-price`)

```json
{
  "request_id": "e16d6d12-b045-4eb9-bf07-b811a3836e57",
  "type": "confirm_price",
  "payload": {
    "order_id": "e16d6d12-b045-4eb9-bf07-b811a3836e57",
    "operator_id": "a1b2c3d4-5e6f-7890-abcd-ef1234567890",
    "accepted_price": 2800.00,
    "commission_amount": 280.00,
    "operator_amount": 2520.00
  }
}
```

---

### Эксплуатант → Агрегатор (`<prefix>.operator.responses`)

#### `price_offer` — эксплуатант называет свою цену

Агрегатор сохраняет `operator_id` и `offered_price` в БД, переводит заказ в `matched`.
Пользователь видит оферту через `GET /orders/{id}`.

```json
{
  "request_id": "e16d6d12-b045-4eb9-bf07-b811a3836e57",
  "type": "price_offer",
  "payload": {
    "order_id": "e16d6d12-b045-4eb9-bf07-b811a3836e57",
    "operator_id": "a1b2c3d4-5e6f-7890-abcd-ef1234567890",
    "operator_name": "ООО АэроДоставка",
    "price": 2800.00,
    "estimated_time_minutes": 25,
    "provided_security_goals": ["ЦБ1"],
    "insurance_coverage": "Лимит 1 млн"
  }
}
```

#### `order_result` — результат выполнения / срыв

Агрегатор переводит заказ в `completed` (success=true) или `dispute` (success=false).

**Успешное выполнение:**

```json
{
  "request_id": "e16d6d12-b045-4eb9-bf07-b811a3836e57",
  "type": "order_result",
  "payload": {
    "order_id": "e16d6d12-b045-4eb9-bf07-b811a3836e57",
    "operator_id": "a1b2c3d4-5e6f-7890-abcd-ef1234567890",
    "success": true,
    "reason": "",
    "total_price": 2800.00
  }
}
```

**Срыв миссии:**

```json
{
  "request_id": "e16d6d12-b045-4eb9-bf07-b811a3836e57",
  "type": "order_result",
  "payload": {
    "order_id": "e16d6d12-b045-4eb9-bf07-b811a3836e57",
    "operator_id": "a1b2c3d4-5e6f-7890-abcd-ef1234567890",
    "success": false,
    "reason": "Потеря связи с дроном на 3-й минуте полёта",
    "total_price": 0
  }
}
```

---

## Топики

Формат имени топика: `<prefix>.<component>.<direction>`, где

- `<prefix> = <protocol_version>.<system_name>.<instance_id>`
- по умолчанию: `v1.aggregator_insurer.local`
- `instance_id` — конкретный ID стенда/команды (например `team42`, `dev-kirill`, `prod-eu1`), чтобы топики не конфликтовали

Это убирает конфликты между экземплярами систем и сразу закладывает версионирование протокола.

| Топик                          | Направление               | Кто читает                        |
| ----------------------------------- | ------------------------------------ | ------------------------------------------ |
| `<prefix>.operator.requests`      | Агрегатор → Эксп.      | Сервис эксплуатанта      |
| `<prefix>.operator.responses`     | Эксп. → Агрегатор      | Агрегатор (этот сервис) |
| `<prefix>.aggregator.requests`    | Внешние → Агрегатор | Агрегатор                         |
| `<prefix>.aggregator.responses`   | Агрегатор → Внешние | Внешние сервисы              |
| `<prefix>.aggregator.dead_letter` | Мусорные сообщения  | —                                         |

---

## Схема базы данных (orders — ключевые поля)

```
customers   — заказчики (id, name, email, phone)
operators   — эксплуатанты (id, name, license, email)
orders      — заказы (id, customer_id→customers, description, budget,
                       mission_type,
                       security_goals[],
                       from_lat/from_lon/to_lat/to_lon (delivery),
                       top_left_lat/top_left_lon/bottom_right_lat/bottom_right_lon (agro),
                       status,
                       operator_id, offered_price,
                       commission_amount, operator_amount,
                       created_at)
```

Миграции применяются автоматически при старте сервиса из файла `migrations/001_init.sql`.
