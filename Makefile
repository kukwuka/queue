MOCK=github.com/vektra/mockery/v2@v2.42.1
LINT=github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.0

.clean-mocks:
	rm -rf ./mocks/

mocks:
	go run $(MOCK) --dir ./internal --all --output ./mocks --with-expecter --keeptree --case snake

lint:
	go run $(LINT) run ./...

format:
	go run $(LINT) cache clean
	go run $(LINT) run --fix ./...

test:
	go test -race ./...

up:
	go run ./cmd -port=8081 -timeout=10s -queueMaxSize=3 -queuesMaxCount=3