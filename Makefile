DIR ?= ..
# VERSION es la única fuente de la semver; `make build` le suma -dev.
VERSION := $(shell cat VERSION)

.PHONY: web build test run

web:
	cd web && npm ci && npm run build

build: web
	go build -ldflags "-X main.version=$(VERSION)-dev" -o bin/sdd ./cmd/sdd

test:
	go vet ./... && go test ./...

run: build
	./bin/sdd --dir "$(DIR)"
