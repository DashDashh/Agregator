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

#### Обновить статус заказа

```
PUT /orders/{id}/status
```

**Тело запроса:**

```json
{"status": "confirmed"}
```

**Доступные статусы:**

| Статус  | Описание                             |
| ------------- | -------------------------------------------- |
| `pending`   | Ждёт исполнителя              |
| `searching` | Идёт поиск эксплуатанта |
| `matched`   | Исполнитель найден          |
| `confirmed` | Контракт подписан            |
| `completed` | Заказ выполнен                  |
| `dispute`   | Открыт спор                        |

**Ответ `200`:**

```json
{"status": "confirmed"}
```

**Ответ `404`:**

```json
{"error": "заказ не найден"}
```

---

## Пример полного запроса

```bash
# 1. Создать заказчика
CUSTOMER_ID=$(curl -s -X POST http://localhost:8080/customers \
  -H "Content-Type: application/json" \
  -d '{"name":"Иван","email":"ivan@mail.ru","phone":"+7900"}' \
  | grep -o '"id":"[^"]*"' | cut -d'"' -f4)

# 2. Создать заказ — сохранится в БД и уйдёт в Kafka
ORDER_ID=$(curl -s -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d "{\"customer_id\":\"$CUSTOMER_ID\",\"description\":\"доставить\",\"budget\":1500}" \
  | grep -o '"id":"[^"]*"' | cut -d'"' -f4)

# 3. Проверить статус (должен быть "pending")
curl -s http://localhost:8080/orders/$ORDER_ID | python3 -m json.tool

# 4. Подтвердить заказ
curl -s -X PUT http://localhost:8080/orders/$ORDER_ID/status \
  -H "Content-Type: application/json" \
  -d '{"status":"confirmed"}'
```

---

## Схема базы данных

```
customers   — заказчики (id, name, email, phone)
operators   — эксплуатанты (id, name, license, email)
orders      — заказы (id, customer_id→customers, description, budget,
                       from_lat, from_lon, to_lat, to_lon, status, created_at)
```

Миграции применяются автоматически при старте сервиса из файла `migrations/001_init.sql`.
