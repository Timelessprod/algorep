BIN=job_scheduler

all: clean build exec

$(BIN): build

build:
	go mod tidy
	go build -o $(BIN) ./cmd/main

exec: $(BIN)
	./$(BIN)

test:
	go test -v

clean:
	rm -f $(BIN)
	rm -f app.log