.PHONY: test build docker run

test:
	go test ./... -v -cover

build:
	CGO_ENABLED=0 go build -o bin/server ./cmd/server

docker:
	docker build -t portfolio-server:local .

run: docker
	docker run --rm -p 8080:8080 portfolio-server:local
