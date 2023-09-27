.PHONY: test
test:
	go test ./... -race -short -v -timeout 60s

.PHONY: build
build: test
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o ./dist/kafka-source main.go

.PHONY: image
image: build
	docker build -t "quay.io/numaio/numaflow-source/kafka-source-go:v0.1.4" --target kafka-source .

clean:
	-rm -rf ./dist