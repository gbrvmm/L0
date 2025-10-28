.PHONY: run publish tidy

BIN_SERVER=orders-server
BIN_PUB=orders-publisher

tidy:
	go mod tidy

build: tidy
	go build -o bin/$(BIN_SERVER) ./cmd/server
	go build -o bin/$(BIN_PUB) ./cmd/publisher

run: tidy
	DB_HOST=localhost DB_PORT=5432 DB_NAME=orders DB_USER=orders DB_PASS=orders 	STAN_CLUSTER_ID=test-cluster STAN_CLIENT_ID=orders-server-1 STAN_URL=nats://localhost:4222 	CHANNEL=orders DURABLE=orders-durable 	go run ./cmd/server

publish: tidy
	STAN_CLUSTER_ID=test-cluster STAN_CLIENT_ID=publisher-1 STAN_URL=nats://localhost:4222 	CHANNEL=orders go run ./cmd/publisher -file ./sample/model.json
