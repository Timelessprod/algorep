BIN=job_scheduler

all: build exec

$(BIN): build

build:
	go build -o $(BIN)

exec: $(BIN)
	./$(BIN)

test:
	go test -v