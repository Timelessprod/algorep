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

testenv:
	docker run -v $$(pwd):/algorep -w /algorep --name algorepenv -it golang:1.18.8
	docker rm algorepenv -f

clean:
	rm -f $(BIN)
	rm -f app.log
	rm -rf state/*.node