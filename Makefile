DIR ?= ..
# VERSION es la única fuente de la semver; `make build` le suma -dev.
VERSION := $(shell cat VERSION)

.PHONY: web build test run release-assets

web:
	cd web && npm ci && npm run build

build: web
	go build -ldflags "-X main.version=$(VERSION)-dev" -o bin/sdd ./cmd/sdd

test:
	go vet ./... && go test ./...

run: build
	./bin/sdd --dir "$(DIR)"

# Binarios del release: uno por plataforma, más checksums.txt.
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64
# Mismo formato que acepta internal/semver.Parse; release.yml corre además
# cmd/versioncheck antes de este target.
SEMVER_RE := ^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$$

# release-assets compila en dist/release los binarios con la VERSION exacta.
# Requiere el front ya buildeado (make web o npm run build en web).
release-assets:
	@printf '%s\n' '$(VERSION)' | grep -Eqx '$(SEMVER_RE)' || { echo "VERSION inválida: '$(VERSION)' no es MAJOR.MINOR.PATCH" >&2; exit 1; }
	@test -f web/dist/index.html || { echo "falta web/dist/index.html: buildeá el front antes (make web)" >&2; exit 1; }
	rm -rf dist/release && mkdir -p dist/release
	@set -e; for p in $(PLATFORMS); do \
		os=$${p%/*}; arch=$${p#*/}; ext=; \
		if [ "$$os" = windows ]; then ext=.exe; fi; \
		out=dist/release/sdd_$(VERSION)_$${os}_$${arch}$$ext; \
		echo "go build $$out"; \
		CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch go build -ldflags "-s -w -X main.version=$(VERSION)" -o $$out ./cmd/sdd; \
	done
	cd dist/release && sha256sum sdd_* > checksums.txt
