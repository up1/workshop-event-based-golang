# Microservice Demo with Watermill and OpenTelemetry

This project demonstrates a microservice architecture using Go, Watermill for message passing, and OpenTelemetry for distributed tracing.

## Architecture

- **Service 1 (Order Service)**: REST API that receives order requests and publishes OrderCreated events
- **Service 2 (Report Service)**: Consumes OrderCreated events and generates reports
- **RabbitMQ**: Message broker using fanout exchange
- **Jaeger**: Distributed tracing system
- **OpenTelemetry**: Observability framework

## Technology Stack

- Go 1.22
- [Watermill](https://watermill.io/) - Event streaming library
- [OpenTelemetry](https://opentelemetry.io/) - Observability framework
- [Jaeger](https://www.jaegertracing.io/) - Distributed tracing
- RabbitMQ - Message broker
- Docker & Docker Compose

## Service Flow

```
Client -> Service1 (REST API) -> RabbitMQ -> Service2 (Event Consumer)
```

## API Documentation

### Create Order API

**Endpoint**: `POST /order`

**Request**:
```json
{
  "total_price": 1000,
  "customer_id": 1,
  "product_id": 1
}
```

**Response** (201 Created):
```json
{
  "order_id": "uuid-string",
  "status": "created"
}
```

### Get Reports API

**Endpoint**: `GET /reports`

**Response** (200 OK):
```json
{
  "reports": [
    {
      "order_id": "uuid-string",
      "total_price": 1000,
      "customer_id": 1,
      "product_id": 1,
      "created_at": "2025-10-30T10:00:00Z",
      "processed_at": "2025-10-30T10:00:01Z"
    }
  ],
  "total": 1
}
```

## Getting Started

### Prerequisites

- Docker
- Docker Compose
- curl (for testing)

### 1. Start the Services

```bash
# Start all services
docker-compose up -d

# Wait for services to be ready (about 30-60 seconds)
docker-compose logs -f
```

### 2. Setup RabbitMQ

```bash
# Run the RabbitMQ setup script
./setup-rabbitmq.sh
```

### 3. Test the Services

```bash
# Run the test script
./test-services.sh
```

Or manually test:

```bash
# Create an order
curl -X POST http://localhost:8080/order \
  -H "Content-Type: application/json" \
  -d '{
    "total_price": 1000,
    "customer_id": 1,
    "product_id": 1
  }'

# Check reports
curl -X GET http://localhost:8081/reports
```

### 4. View Tracing Data

- **Jaeger UI**: http://localhost:16686
- **RabbitMQ Management**: http://localhost:15672 (guest/guest)

## Distributed Tracing

The application implements distributed tracing using OpenTelemetry:

1. **Service1** creates a trace when receiving HTTP requests
2. **Trace context** is propagated through RabbitMQ message headers
3. **Service2** extracts the trace context and continues the trace
4. **Jaeger** collects and displays the complete trace

### Trace Flow

1. HTTP Request → Service1 (creates span)
2. Service1 → RabbitMQ (publishes message with trace context)
3. RabbitMQ → Service2 (consumes message and extracts trace context)
4. Service2 processes event (continues trace)

## Development

### Building Services Locally

```bash
# Build service1
cd service1
go mod tidy
go build -o service1 .

# Build service2
cd ../service2
go mod tidy
go build -o service2 .
```

### Running Services Locally

You'll need to start RabbitMQ and Jaeger first:

```bash
# Start only infrastructure
docker-compose up -d rabbitmq jaeger

# Run services locally
cd service1
./service1 &

cd ../service2
./service2 &
```

## Troubleshooting

### Common Issues

1. **Services can't connect to RabbitMQ**
   - Ensure RabbitMQ is fully started before starting services
   - Check RabbitMQ logs: `docker-compose logs rabbitmq`

2. **No traces in Jaeger**
   - Verify Jaeger is accessible at http://localhost:16686
   - Check service logs for OpenTelemetry errors

3. **Messages not being processed**
   - Verify RabbitMQ setup with: `./setup-rabbitmq.sh`
   - Check queue bindings in RabbitMQ management UI

### Useful Commands

```bash
# View logs
docker-compose logs -f service1
docker-compose logs -f service2

# Restart a service
docker-compose restart service1

# View RabbitMQ queue status
docker exec rabbitmq rabbitmqadmin list queues

# Clean up
docker-compose down -v
```

## Features Demonstrated

- ✅ REST API with Gin
- ✅ Event publishing with Watermill
- ✅ Event consumption with Watermill
- ✅ RabbitMQ fanout exchange
- ✅ Distributed tracing with OpenTelemetry
- ✅ Jaeger integration
- ✅ Docker containerization
- ✅ Health check endpoints
- ✅ Graceful error handling
- ✅ Message acknowledgment
- ✅ Trace context propagation

## License

This project is for demonstration purposes.