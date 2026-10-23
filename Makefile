DIR ?= ..

.PHONY: web build test run

web:
	cd web && npm ci && npm run build

build: web
	go build -o bin/sdd ./cmd/sdd

test:
	go vet ./... && go test ./...

run: build
	./bin/sdd --dir "$(DIR)"
