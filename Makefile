APP := configaudit

.PHONY: fmt tidy test build run clean docker-build

fmt:
	gofmt -w .

tidy:
	go mod tidy

test:
	go test ./...

build:
	go build -o $(APP) ./cmd/configaudit

docker-build:
	docker build -t $(APP) .

run:
	go run ./cmd/configaudit testdata/clean.yaml

clean:
	rm -f $(APP)
