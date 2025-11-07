BIN_DIR=bin

all: build

build:
	go build -o $(BIN_DIR)

test:
	go test ./...

clean:
	rm -rf $(BIN_DIR)/*
