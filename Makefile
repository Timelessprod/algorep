BIN=job_scheduler

all: build exec

$(BIN): build

build:
	go mod tidy
	go build -o $(BIN)

exec: $(BIN)
	./$(BIN) 2>&1 | tee app.log

test:
	go test -v