.DEFAULT_GOAL := build

GO ?= go
NPM ?= npm

.PHONY: build frontend run dev-api dev-web test test-e2e

frontend:
	cd web && $(NPM) ci --no-audit --no-fund && $(NPM) run build

build: frontend
	mkdir -p bin
	CGO_ENABLED=0 $(GO) build -trimpath -o bin/countdown-studio .

run: build
	./bin/countdown-studio

dev-api:
	$(GO) run .

dev-web:
	cd web && $(NPM) run dev

test:
	$(GO) test ./...
	cd web && $(NPM) test

test-e2e: build
	cd web && $(NPM) run test:e2e
