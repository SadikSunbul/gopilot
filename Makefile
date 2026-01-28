SHELL := /bin/bash

GO ?= go
GOLANGCI_LINT ?= golangci-lint

.PHONY: help fmt vet tidy test test-race coverage lint build run run-cli clean

help:
	@echo "Kullanım: make <hedef>"
	@echo ""
	@echo "Hedefler:"
	@echo "  fmt        Go formatla (gofmt)"
	@echo "  vet        go vet"
	@echo "  tidy       go mod tidy"
	@echo "  test       testleri çalıştır"
	@echo "  test-race  testleri race ile çalıştır"
	@echo "  coverage   coverage üret (coverage.out)"
	@echo "  lint       golangci-lint çalıştır (yoksa kurulum komutu gösterir)"
	@echo "  build      tüm paketleri build et"
	@echo "  run        offline örneği çalıştır (API key gerekmez)"
	@echo "  run-cli    interactive CLI örneğini çalıştır (GEMINI_API_KEY gerekli)"
	@echo "  clean      build/coverage artefact'lerini sil"

fmt:
	@$(GO) fmt ./...
	@gofmt -w .

vet:
	@$(GO) vet ./...

tidy:
	@$(GO) mod tidy

test:
	@$(GO) test -v ./...

test-race:
	@$(GO) test -v -race ./...

coverage:
	@$(GO) test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@$(GO) tool cover -func=coverage.out | tail -n 1 || true

lint:
	@if ! command -v "$(GOLANGCI_LINT)" >/dev/null 2>&1; then \
		echo "golangci-lint bulunamadı. Kurmak için:"; \
		echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi
	@$(GOLANGCI_LINT) run --timeout=5m ./...

build:
	@$(GO) build -v ./...

run:
	@$(GO) run ./examples/offline

run-cli:
	@$(GO) run ./examples/cli

clean:
	@rm -f coverage.out
