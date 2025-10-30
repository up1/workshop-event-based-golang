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

$go test product_service_test.go --count=1 -v -cover

$go run cmd/server/main.go
```

Go to swagger 
* http://localhost:8080/swagger/index.html


## 2. Run consumer
```
$cd consumer1
$go mod tidy

$go test client_test.go --count=1 -v -cover

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

Install `pact-go` CLI and Pact FFI
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

## 5. Publish contract to Pact Broker from consumer-side
* Use [pact-broker-client](https://github.com/pact-foundation/pact-standalone/releases) CLI

Install
```
$cd consumer1

$curl -fsSL https://raw.githubusercontent.com/pact-foundation/pact-standalone/master/install.sh | PACT_CLI_VERSION=v2.5.5 bash

$export PATH=.:/Users/somkiatpuisungnoen/data/slide/microservice/demo-dime/contract_test/consumer1/pact/bin/:${PATH}

$pact-broker version
```

Publish
```

$export PACT_BROKER_PROTO=http
$export PACT_BROKER_URL=localhost:9292
$export VERSION_COMMIT=1.0
$export PACT_BROKER_USERNAME=pact_workshop
$export PACT_BROKER_PASSWORD=pact_workshop


$pact-broker publish ${PWD}/pacts --consumer-app-version ${VERSION_COMMIT} --branch ${VERSION_BRANCH} -b ${PACT_BROKER_PROTO}://${PACT_BROKER_URL} -u ${PACT_BROKER_USERNAME} -p ${PACT_BROKER_PASSWORD}
```

Check contract in Pact broker
* http://localhost:9292/


## 6. Verify contract in provider-side
```
$cd provider1

$export PACT_BROKER_PROTO=http
$export PACT_BROKER_URL=localhost:9292
$export VERSION_COMMIT=1.0
$export PACT_BROKER_USERNAME=pact_workshop
$export PACT_BROKER_PASSWORD=pact_workshop
$go test product_service_pact_test.go
```
