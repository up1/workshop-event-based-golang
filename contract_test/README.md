# Contract testing workshop with Go
* [Pact Go](https://github.com/pact-foundation/pact-go)
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

## 4. Create contract with Pact (Consumer-side)



Install `pact-go` CLI and Pact FFI on MacOS
```
$go install github.com/pact-foundation/pact-go/v2
$pact-go version

$sudo pact-go install
$otool -L /usr/local/lib/libpact_ffi.dylib
```

Run test to create Pact's contract file
```
$cd consumer1

$export LOG_LEVEL=info 
$export CONSUMER_NAME=consumer1
$export PROVIDER_NAME=provider1

$go test client_pact_test.go -v
$go test -tags=integration -count=1 ./... -run 'TestClientPact' -v
```

Open contract file
* /pacts/consumer1-provider1.json