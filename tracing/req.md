## Order service and Report service
* 1. Client sent data to service1 with REST API POST /order
* 2. Service1 send or publish a message(OrderCreated Event) to rabbitmq server (exchange=orders, exchange_type=fanout)
* 3. Service2 receive or subscribe messages(OrderCreated Event) from rabbitmq server (exchange=orders, exchange_type=fanout, queue=report)

## Technology stack
* go 1.25.0
* Working with watermill https://watermill.io/
* Observability with distributed tracing with opentelemetry (service1 -> rabbitmq -> service2)
* Opentelemetry and Jaeger https://www.jaegertracing.io/
* Working with docker compose

## Service flow
* service1 -> rabbitmq -> service2

## Create a new Order API

### POST /order

Request
```
POST /order
Content-Type: application/json
{
  "total_proce": 1000,
  "customer_id": 1,
  "product_id": 1
}
```

Success Response with code= 201
```
{
  "order_id": "0001",
  "status": "created"
}
```

## OrderCreated Event with tracing from opentelemetry
{
  "order_id": "0001",
  "total_proce": 1000,
  "customer_id": 1,
  "product_id": 1
}