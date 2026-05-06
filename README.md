# Agregator

Агрегатор страховых заявок на доставку дронами.

Проект можно запускать в двух режимах:

- обычный режим: gateway вызывает серверные компоненты внутри одного процесса;
- микросервисный режим: gateway принимает HTTP-запросы и сообщения Kafka, а `registry`, `orders`, `contracts`, `analytics` запускаются отдельными сервисами и обмениваются сообщениями через Kafka.

## Стек

- Go 1.24
- PostgreSQL 16
- Apache Kafka
- MQTT для дополнительного обмена с эксплуатантами
- Docker Compose

## Архитектура

Основные серверные компоненты:

| Компонент | Назначение | Запуск |
| --- | --- | --- |
| `gateway` | HTTP API, фронтенд, входной системный топик Kafka, маршрутизация | `src/gateway` |
| `registry` | заказчики, эксплуатанты, авторизация | `cmd/registry` |
| `orders` | создание заказов и подбор исполнителя | `cmd/orders` |
| `contracts` | цена, подтверждение, выполнение, спор | `cmd/contracts` |
| `analytics` | аналитика | `cmd/analytics` |
| `operator_exchange` | обмен с внешним сервисом эксплуатантов через Kafka/MQTT | `src/operator_exchange_component` |

Общий код находится в `src/shared`: доменные типы, модели сообщений, ответы, адаптер PostgreSQL, Kafka-настройки и общая шина компонентов.

Фронтенд, PostgreSQL, Kafka и MQTT считаются внешней инфраструктурой, а не доверенными серверными компонентами.

## Режимы запуска

### Обычный локальный режим с Kafka

```bash
docker network create drones_net
make docker-up-dev
```

Если сеть уже существует, сообщение `network with name drones_net already exists` можно игнорировать.

В этом режиме поднимаются:

- `aggregator`
- `postgres`
- `zookeeper`
- `kafka`
- `kafka-init`

Gateway работает с компонентами внутри процесса. Это режим совместимости.

### Микросервисный режим

```bash
docker network create drones_net
make docker-up-micro
```

В этом режиме поднимаются:

- `aggregator`
- `registry`
- `orders`
- `contracts`
- `analytics`
- `postgres`
- `zookeeper`
- `kafka`
- `kafka-init`

`kafka-init` создает нужные топики и завершается с кодом `0`. Это нормально.

Проверить контейнеры:

```bash
docker compose -f docker-compose.yml -f docker-compose.dev.yml --profile kafka --profile microservices ps -a
```

Ожидаемо:

- `aggregator` — `Up`
- `registry` — `Up`
- `orders` — `Up`
- `contracts` — `Up`
- `analytics` — `Up`
- `postgres` — `Up (healthy)`
- `zookeeper` — `Up`
- `kafka` — `Up`
- `kafka-init` — `Exited (0)`

## Проверка работоспособности

### Бэкенд

```bash
curl http://localhost:8081/health
```

Ожидаемый ответ:

```json
{"status":"ok"}
```

### Фронтенд

Открыть в браузере:

```text
http://localhost:8081
```

Frontend отдается gateway из папки `frontend`.

### Логи

Для микросервисного режима:

```bash
docker compose -f docker-compose.yml -f docker-compose.dev.yml --profile kafka --profile microservices logs -f aggregator registry orders contracts analytics kafka
```

Для обычного режима разработки:

```bash
docker compose -f docker-compose.yml -f docker-compose.dev.yml --profile kafka logs -f aggregator kafka
```

## Быстрая проверка через HTTP

### 1. Зарегистрировать заказчика

```bash
curl -X POST http://localhost:8081/customers \
  -H "Content-Type: application/json" \
  -d '{"name":"Иван","email":"ivan@example.com","password":"strongpass123","phone":"+79001234567"}'
```

Скопировать `token` и `user.id` из ответа:

```bash
CUSTOMER_TOKEN='сюда_токен_заказчика'
CUSTOMER_ID='сюда_id_заказчика'
```

### 2. Создать заказ

```bash
curl -X POST http://localhost:8081/orders \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $CUSTOMER_TOKEN" \
  -d '{"description":"Доставка документов","budget":1000,"from_lat":55.75,"from_lon":37.61,"to_lat":55.76,"to_lon":37.62,"mission_type":"delivery"}'
```

Скопировать `id` заказа:

```bash
ORDER_ID='сюда_id_заказа'
```

### 3. Зарегистрировать эксплуатанта

```bash
curl -X POST http://localhost:8081/operators \
  -H "Content-Type: application/json" \
  -d '{"name":"Оператор 1","email":"operator@example.com","password":"strongpass123","license":"LIC-001"}'
```

Скопировать `token` и `user.id`:

```bash
OPERATOR_TOKEN='сюда_токен_эксплуатанта'
OPERATOR_ID='сюда_id_эксплуатанта'
```

### 4. Предложить цену

```bash
curl -X POST http://localhost:8081/orders/$ORDER_ID/offer \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $OPERATOR_TOKEN" \
  -d '{"price":900}'
```

### 5. Подтвердить цену заказчиком

```bash
curl -X POST http://localhost:8081/orders/$ORDER_ID/confirm-price \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $CUSTOMER_TOKEN" \
  -d '{"operator_id":"'"$OPERATOR_ID"'","accepted_price":900}'
```

### 6. Проверить заказ

```bash
curl http://localhost:8081/orders/$ORDER_ID \
  -H "Authorization: Bearer $CUSTOMER_TOKEN"
```

## Тесты и сборка

Запуск всех Go-тестов с покрытием:

```bash
GOCACHE=/tmp/go-build go test ./... -cover
```

Сборка gateway и компонентных сервисов:

```bash
GOCACHE=/tmp/go-build go build ./src/gateway ./cmd/registry ./cmd/orders ./cmd/contracts ./cmd/analytics
```

Проверка Docker Compose конфигурации:

```bash
docker compose -f docker-compose.yml -f docker-compose.dev.yml --profile kafka --profile microservices config
```

Сборка Docker-образов:

```bash
docker compose -f docker-compose.yml -f docker-compose.dev.yml --profile kafka --profile microservices build aggregator registry orders contracts analytics
```

## Переменные окружения

| Переменная | Значение по умолчанию | Назначение |
| --- | --- | --- |
| `COMPONENT_DISPATCH_MODE` | `inprocess` | `inprocess` или `broker`; режим маршрутизации между gateway и компонентами |
| `KAFKA_BROKER` | `localhost:9092` / `kafka:9092` в dev compose | адрес Kafka |
| `SYSTEM_NAMESPACE` | пусто | префикс для системных топиков |
| `OPERATOR_TRANSPORT` | `kafka` | `kafka` или `both`; обмен с эксплуатантами |
| `AUTH_SECRET` | dev-значение | секрет для токенов |
| `DATABASE_URL` | dev-значение | подключение к PostgreSQL |
| `COMMISSION_RATE` | `0.1` | комиссия агрегатора |

## Режим транспорта

Есть два независимых уровня транспорта.

`COMPONENT_DISPATCH_MODE` управляет связью между gateway и серверными компонентами:

- `inprocess` — gateway вызывает обработчики компонентов внутри своего процесса;
- `broker` — gateway отправляет запросы в Kafka topics отдельных компонентов.

Для запуска через брокер используйте:

```bash
make docker-up-micro
```

`OPERATOR_TRANSPORT` управляет обменом с внешним сервисом эксплуатантов через `operator.requests` и `operator.responses`:

- `kafka` — только Kafka, режим по умолчанию;
- `both` — Kafka и MQTT одновременно.

Пример запуска с Kafka + MQTT для эксплуатантов:

```bash
OPERATOR_TRANSPORT=both make docker-up-micro
```

Системный входной топик и топики ответов могут получать префикс `SYSTEM_NAMESPACE`. Если `SYSTEM_NAMESPACE=fleet_1`, то `systems.agregator` превращается в `fleet_1.systems.agregator`.

## Топики Kafka

Основные топики агрегатора:

| Топик | Кто читает |
| --- | --- |
| `systems.agregator` | gateway |
| `components.agregator.responses` | внешние системы |
| `components.agregator.registry` | сервис registry |
| `components.agregator.orders` | сервис orders |
| `components.agregator.contracts` | сервис contracts |
| `components.agregator.analytics` | сервис analytics |
| `components.agregator.operator.requests` | сервис эксплуатантов |
| `components.agregator.operator.responses` | агрегатор |
| `errors.dead_letters` | мониторинг ошибок |

В микросервисном режиме gateway получает сообщение из `systems.agregator`, определяет компонент по `action` и отправляет запрос в один из `components.agregator.*`.

## Форматы сообщений

В межсистемном взаимодействии отправитель публикует сообщение в системный топик получателя, например `systems.agregator`. Gateway читает системный топик, смотрит поле `action` и маршрутизирует запрос в нужный компонент.

### Входящий запрос

```json
{
  "action": "create_order",
  "payload": {
    "customer_id": "customer-1",
    "description": "Доставка документов",
    "budget": 3000,
    "mission_type": "delivery",
    "from_lat": 55.7558,
    "from_lon": 37.6173,
    "to_lat": 55.8,
    "to_lon": 37.65
  },
  "sender": "external_system",
  "correlation_id": "order-1",
  "reply_to": "components.agregator.responses",
  "timestamp": "2026-05-06T12:00:00Z"
}
```

Вместо `action` также поддерживается поле `type`, а вместо `correlation_id` — `request_id`. Это сделано для совместимости с разными отправителями.

### Ответ

```json
{
  "action": "response",
  "payload": {
    "order_id": "order-1",
    "status": "pending",
    "message": "order created, awaiting executor selection (stub)"
  },
  "sender": "agregator",
  "correlation_id": "order-1",
  "success": true,
  "timestamp": "2026-05-06T12:00:01Z"
}
```

При ошибке `success` будет `false`, а причина будет в поле `error`.

### Сообщение агрегатора эксплуатанту

При создании заказа агрегатор публикует сообщение в `components.agregator.operator.requests`:

```json
{
  "action": "create_order",
  "sender": "agregator",
  "correlation_id": "order-1",
  "payload": {
    "customer_id": "customer-1",
    "description": "Доставка документов",
    "budget": 3000,
    "mission_type": "delivery",
    "security_goals": ["ЦБ1"],
    "from_lat": 55.7558,
    "from_lon": 37.6173,
    "to_lat": 55.8,
    "to_lon": 37.65
  }
}
```

После подтверждения цены агрегатор отправляет эксплуатанту `confirm_price`:

```json
{
  "action": "confirm_price",
  "sender": "agregator",
  "correlation_id": "order-1",
  "payload": {
    "order_id": "order-1",
    "operator_id": "operator-1",
    "accepted_price": 2800,
    "commission_amount": 280,
    "operator_amount": 2520
  }
}
```

### Сообщение эксплуатанта агрегатору

Эксплуатант пишет ответы в `components.agregator.operator.responses`.

Оферта цены:

```json
{
  "action": "price_offer",
  "sender": "operator_service",
  "correlation_id": "order-1",
  "payload": {
    "order_id": "order-1",
    "operator_id": "operator-1",
    "operator_name": "Оператор 1",
    "price": 2800,
    "estimated_time_minutes": 25,
    "provided_security_goals": ["ЦБ1"],
    "insurance_coverage": "Лимит 1 млн"
  }
}
```

Результат выполнения:

```json
{
  "action": "order_result",
  "sender": "operator_service",
  "correlation_id": "order-1",
  "payload": {
    "order_id": "order-1",
    "operator_id": "operator-1",
    "success": true,
    "reason": "",
    "total_price": 2800
  }
}
```

Отправить тестовое сообщение в Kafka можно так:

```bash
docker compose -f docker-compose.yml -f docker-compose.dev.yml --profile kafka exec -T kafka \
  kafka-console-producer \
  --bootstrap-server kafka:9092 \
  --topic components.agregator.operator.responses <<EOF
{"action":"price_offer","sender":"operator_service","correlation_id":"$ORDER_ID","payload":{"order_id":"$ORDER_ID","operator_id":"$OPERATOR_ID","operator_name":"Оператор 1","price":900,"estimated_time_minutes":25,"provided_security_goals":["ЦБ1"],"insurance_coverage":"Лимит 1 млн"}}
EOF
```

## Основные HTTP ручки

| Метод | Путь | Назначение |
| --- | --- | --- |
| `GET` | `/health` | проверка сервиса |
| `POST` | `/customers` | регистрация заказчика |
| `POST` | `/operators` | регистрация эксплуатанта |
| `POST` | `/auth/login` | вход |
| `POST` | `/orders` | создание заказа |
| `GET` | `/orders` | список заказов |
| `GET` | `/orders/{id}` | заказ по id |
| `POST` | `/orders/{id}/offer` | предложение цены эксплуатантом |
| `POST` | `/orders/{id}/confirm-price` | подтверждение цены заказчиком |
| `POST` | `/orders/{id}/confirm-completion` | подтверждение выполнения |

Защищенные ручки требуют заголовок:

```http
Authorization: Bearer <token>
```

## Статусы заказа

| Статус | Когда выставляется |
| --- | --- |
| `pending` | заказ создан в БД |
| `searching` | заказ опубликован эксплуатантам |
| `matched` | эксплуатант предложил цену |
| `confirmed` | заказчик подтвердил цену |
| `completed_pending_confirmation` | эксплуатант сообщил об успешном выполнении |
| `completed` | заказчик подтвердил выполнение |
| `dispute` | эксплуатант сообщил о срыве |

## Остановка

Остановить контейнеры без удаления данных БД:

```bash
docker compose -f docker-compose.yml -f docker-compose.dev.yml --profile kafka --profile microservices down
```

Остановить и удалить volume с данными:

```bash
docker compose -f docker-compose.yml -f docker-compose.dev.yml --profile kafka --profile microservices down -v
```

Для обычной работы используйте остановку без `-v`.

## Миграции

Миграции применяются gateway при старте из файла:

```text
migrations/001_init.sql
```
