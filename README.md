# Agregator Insurer

Агрегатор страховых заявок на доставку дронами.

## Стек

- **Go 1.24** — сервис
- **PostgreSQL 16** — хранение заказчиков, эксплуатантов и заказов
- **Apache Kafka** — асинхронная очередь заданий для эксплуатантов
- **Docker Compose** — запуск всего окружения одной командой

## Запуск

```bash
docker compose up -d --build
```

Сервис поднимется на `http://localhost:8080`.

Порядок старта автоматический: сначала PostgreSQL (с healthcheck), затем Kafka, затем агрегатор.

---

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

> Требует существующего `customer_id`. При создании заказ автоматически отправляется эксплуатантам через Kafka (`operator.requests`).

**Тело запроса:**

```json
{
  "customer_id": "b9e8b4d6-2318-429b-944c-11e46db1fbfe",
  "description": "Доставить документы из офиса на склад",
  "budget": 2500.00,
  "from_lat": 55.7558,
  "from_lon": 37.6173,
  "to_lat": 55.8000,
  "to_lon": 37.6500
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
> и отправляет эксплуатанту сообщение `confirm_price` через Kafka (`operator.requests`).

**Тело запроса:**

```json
{
  "operator_id": "a1b2c3d4-...",
  "accepted_price": 2200.00
}
```

**Ответ `200`:**

```json
{
  "order_id": "e16d6d12-...",
  "operator_id": "a1b2c3d4-...",
  "accepted_price": 2200.00,
  "status": "confirmed"
}
```

---

**Статусы заказа:**

| Статус      | Когда выставляется                                         |
|-------------|-------------------------------------------------------------|
| `pending`   | Заказ создан, ждёт предложений                              |
| `searching` | Агрегатор опубликовал заказ в Kafka (`operator.requests`)   |
| `matched`   | Эксплуатант прислал оферту цены (`price_offer`)             |
| `confirmed` | Пользователь принял цену (`POST .../confirm-price`)         |
| `completed` | Эксплуатант сообщил об успешном выполнении (`order_result`) |
| `dispute`   | Эксплуатант сообщил о срыве (`order_result` success=false)  |

---

## Пример полного запроса

```bash
# 1. Создать заказчика
CUSTOMER_ID=$(curl -s -X POST http://localhost:8080/customers \
  -H "Content-Type: application/json" \
  -d '{"name":"Иван","email":"ivan@mail.ru","phone":"+7900"}' \
  | grep -o '"id":"[^"]*"' | cut -d'"' -f4)

# 2. Создать заказ — сохранится в БД и уйдёт в Kafka (operator.requests)
ORDER_ID=$(curl -s -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d "{\"customer_id\":\"$CUSTOMER_ID\",\"description\":\"доставить\",\"budget\":3000}" \
  | grep -o '"id":"[^"]*"' | cut -d'"' -f4)

# 3. Проверить статус (pending → searching после отправки в Kafka)
curl -s http://localhost:8080/orders/$ORDER_ID | python3 -m json.tool

# 4. После того как эксплуатант прислал оферту (price_offer через operator.responses),
#    статус стал "matched", в ответе появились offered_price и operator_id.
#    Пользователь принимает цену:
curl -s -X POST http://localhost:8080/orders/$ORDER_ID/confirm-price \
  -H "Content-Type: application/json" \
  -d "{\"operator_id\":\"$OPERATOR_ID\",\"accepted_price\":2800}"
```

---

## Kafka — форматы сообщений

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
    "from_lat": 55.7558,
    "from_lon": 37.6173,
    "to_lat": 55.8000,
    "to_lon": 37.6500
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
    "accepted_price": 2800.00
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
    "estimated_time_minutes": 25
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
    "reason": ""
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
    "reason": "Потеря связи с дроном на 3-й минуте полёта"
  }
}
```

---

## Kafka — топики

Формат имени топика: `<prefix>.<component>.<direction>`, где

- `<prefix> = <protocol_version>.<system_name>.<instance_id>`
- по умолчанию: `v1.aggregator_insurer.local`
- для разных стендов меняется `instance_id` (например `dev-kirill`, `test-team4`, `prod-eu1`)

Это убирает конфликты между экземплярами систем и сразу закладывает версионирование протокола.

| Топик | Направление | Кто читает |
|-------------------------|---------------------|--------------------|
| `<prefix>.operator.requests` | Агрегатор → Эксп. | Сервис эксплуатанта |
| `<prefix>.operator.responses` | Эксп. → Агрегатор | Агрегатор (этот сервис) |
| `<prefix>.aggregator.requests` | Внешние → Агрегатор | Агрегатор |
| `<prefix>.aggregator.responses` | Агрегатор → Внешние | Внешние сервисы |
| `<prefix>.aggregator.dead_letter` | Мусорные сообщения | — |

---

## Схема базы данных

```
customers   — заказчики (id, name, email, phone)
operators   — эксплуатанты (id, name, license, email)
orders      — заказы (id, customer_id→customers, description, budget,
                       from_lat, from_lon, to_lat, to_lon,
                       status, operator_id, offered_price, created_at)
```

Миграции применяются автоматически при старте сервиса из файла `migrations/001_init.sql`.