# Contract testing workshop with Go
* Provider = product service
  * REST API
* Consumer
  * Command line app

## 1. Start with provider
```
$cd provider1
$go mod tidy

$go test ./... --count=1 -v -cover

$go run cmd/server/main.go
```

Go to swagger 
* http://localhost:8080/swagger/index.html


## 2. Run consumer
```
$cd consumer1
$go mod tidy

$go test ./... --count=1 -v -cover

$go run cmd/main.go
```

## 3. Start Pact broker
```
$docker compose up -d
$docker compose ps
```

Access to pact broker
* http://localhost:9292
  * user=pact_workshop
  * pass=pact_workshop

## Verify with contract from Pact broker