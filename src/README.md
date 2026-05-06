# Component Layout

Trusted backend code is grouped by component ownership under `src`.

## Components

- `gateway` - process entrypoint, HTTP router, middleware, config, internal bus routing.
- `registry_component` - customer/operator registration, login, password hashing, token handling.
- `orders_component` - order creation, order lookup, executor search commands.
- `contracts_component` - price confirmation, completion confirmation, dispute/contract commands.
- `analytics_component` - analytics command handling.
- `operator_exchange_component` - trusted adapters for external operator messaging over Kafka/MQTT.

## Shared Code

`shared` is not a runtime component. It contains contracts and adapters that are still shared while
the project runs as one Go process:

- `shared/models` - message and payload contracts.
- `shared/response` - bus response helpers.
- `shared/httpx` - HTTP JSON response helpers.
- `shared/store` - PostgreSQL persistence adapter.

Domain HTTP handlers depend on small local interfaces instead of a concrete `*store.Store`.
This keeps the current behavior intact while making the next split into separate services smaller.
