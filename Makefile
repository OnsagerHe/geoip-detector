all: build
.PHONY: all

build:
	go build -o bin/geoip-detector main.go
.PHONY: build

run:
	go run main.go
.PHONY: run

clean:
	rm -r bin
.PHONY: clean
