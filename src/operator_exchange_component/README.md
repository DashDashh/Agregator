# Operator Exchange Component

This component owns the trusted code that exchanges messages with external drone operators.

Kafka and MQTT brokers are external infrastructure and are not part of this component's trusted
code. The trusted code here is the adapter logic that:

- publishes order requests to operators;
- publishes confirmed prices to operators;
- consumes `price_offer` messages;
- consumes `order_result` messages;
- updates order state through the persistence adapter.

The component currently has Kafka and MQTT backends:

- `kafka`
- `mqtt`

Both backends implement the same publishing behavior expected by the gateway publisher fan-out.
