BIN=job_scheduler

all: build exec

$(BIN): build

build:
	go mod tidy
	go build -o $(BIN) ./cmd/main

exec: $(BIN)
	./$(BIN)

test:
	go test -v