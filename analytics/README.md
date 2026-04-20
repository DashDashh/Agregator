# Analytics module

Минимальная реализация ОФ9: недоверенный read-only модуль аналитики на Python/Pandas.

## Назначение

Модуль читает снимок данных из PostgreSQL и формирует JSON-отчёт для оператора платформы:

- экономические метрики: GMV, средний чек, доход платформы по комиссиям;
- операционные метрики: количество заказов, активные заказчики/эксплуатанты, конверсия из поиска в контракт;
- security-метрики: количество споров, простые antifraud/anomaly alerts, надёжность исполнителей;
- технические метрики: нагрузка по часам и точки для тепловой карты полётов.

## Кибериммунное ограничение

Аналитика считается untrusted-слоем:

- подключение к БД открывается в read-only режиме;
- модуль не пишет в таблицы `orders`, `customers`, `operators`;
- модуль не вызывает управляющие API агрегатора;
- поток данных направлен только в сторону аналитики.

## Запуск

```bash
docker network create drones_net
docker compose -f docker-compose.yml -f docker-compose.dev.yml --profile kafka up -d --build
```

Основной способ запуска аналитики - через временный Python-контейнер в той же Docker-сети:

```bash
docker run --rm \
  --network drones_net \
  -v "$PWD":/app \
  -w /app \
  python:3.11-slim \
  sh -c "pip install -r analytics/requirements.txt && ANALYTICS_DATABASE_URL='postgres://aggregator:secret@postgres:5432/aggregator?sslmode=disable' python analytics/report.py"
```

Так зависимости устанавливаются только внутри временного контейнера, а не на хостовую машину.

Чтобы сохранить отчёт в файл:

```bash
docker run --rm \
  --network drones_net \
  -v "$PWD":/app \
  -w /app \
  python:3.11-slim \
  sh -c "pip install -r analytics/requirements.txt && ANALYTICS_DATABASE_URL='postgres://aggregator:secret@postgres:5432/aggregator?sslmode=disable' python analytics/report.py --output analytics/report_output.json"
```

Если PostgreSQL дополнительно проброшен на `localhost:5432`, можно запускать и из локального виртуального окружения:

```bash
python3 -m venv .venv
source .venv/bin/activate
pip install -r analytics/requirements.txt
ANALYTICS_DATABASE_URL="postgres://aggregator:secret@localhost:5432/aggregator?sslmode=disable" \
  python analytics/report.py
```

## Первый шаг

Это базовый ETL/reporting слой. Следующий логичный шаг - подключить его к отдельной read-only реплике или Kafka-потоку событий и добавить API для фронтенда оператора.