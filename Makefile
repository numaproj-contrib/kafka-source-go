.PHONY: test
test:
	go test ./... -race -short -v -timeout 60s

.PHONY: build
build: test
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o ./dist/kafka-source main.go

.PHONY: image
image: build
	docker build -t "quay.io/numaio/numaflow-source/kafka-source-go:v0.1.8" --target kafka-source .

$(GOPATH)/bin/golangci-lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b `go env GOPATH`/bin v1.54.1

.PHONY: lint
lint: $(GOPATH)/bin/golangci-lint
	go mod tidy
	golangci-lint run --fix --verbose --concurrency 4 --timeout 5m

clean:
	-rm -rf ./dist