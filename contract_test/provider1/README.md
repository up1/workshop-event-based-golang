# Provider :: Product Service

This is a sample product service provider with Swagger/OpenAPI documentation.

## Features

- REST API for managing products
- Swagger/OpenAPI documentation  
- Gorilla Mux router
- JSON responses

## API Endpoints

- `GET /api/v1/products` - Get all products
- `GET /api/v1/products/{id}` - Get product by ID

## Running the Service

1. Install dependencies:
   ```bash
   go mod tidy
   ```

2. Run the server:
   ```bash
   go run cmd/server/main.go
   ```

3. The service will start on port 8080

## Swagger Documentation

Once the service is running, you can access the Swagger UI at:
- **Swagger UI**: http://localhost:8080/swagger/index.html

The API documentation is also available in multiple formats:
- **JSON**: Available at runtime via the API
- **YAML**: Available in `docs/swagger.yaml`

## Generate Swagger Documentation

If you modify the API endpoints or add new ones, regenerate the Swagger documentation:

```bash
# Install swag if not already installed
go install github.com/swaggo/swag/cmd/swag@latest

# Generate docs
swag init -g cmd/server/main.go
```

## Testing
```
$go fmt ./...
$go test ./... --count=1 -v -cover
```

## Verify with Pact

The service maintains compatibility with Pact contract testing