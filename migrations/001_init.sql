-- Заказчики (пользователи, которые создают заказы)
CREATE TABLE IF NOT EXISTS customers (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    email       TEXT NOT NULL,
    phone       TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Эксплуатанты (компании которые выполняют доставку дронами)
CREATE TABLE IF NOT EXISTS operators (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    license     TEXT NOT NULL,
    email       TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Заказы на доставку
CREATE TABLE IF NOT EXISTS orders (
    id            TEXT PRIMARY KEY,
    customer_id   TEXT NOT NULL REFERENCES customers(id),
    description   TEXT NOT NULL,
    budget        NUMERIC(12, 2) NOT NULL DEFAULT 0,
    from_lat      DOUBLE PRECISION NOT NULL DEFAULT 0,
    from_lon      DOUBLE PRECISION NOT NULL DEFAULT 0,
    to_lat        DOUBLE PRECISION NOT NULL DEFAULT 0,
    to_lon        DOUBLE PRECISION NOT NULL DEFAULT 0,
    status        TEXT NOT NULL DEFAULT 'pending',
    operator_id   TEXT NOT NULL DEFAULT '',          -- заполняется когда эксплуатант даёт оферту
    offered_price NUMERIC(12, 2) NOT NULL DEFAULT 0, -- цена предложенная эксплуатантом
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- На случай уже существующей таблицы без новых колонок — добиваем их идемпотентно
ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS operator_id TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS offered_price NUMERIC(12, 2) NOT NULL DEFAULT 0;

-- Индекс для быстрого поиска заказов по заказчику
CREATE INDEX IF NOT EXISTS idx_orders_customer_id ON orders(customer_id);
-- Индекс для фильтрации по статусу
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
