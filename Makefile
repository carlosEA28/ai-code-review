SHELL := /bin/bash

GOOSE_VERSION := v3.27.0
GO_BIN_DIR := $(shell sh -c 'gobin=$$(go env GOBIN); if [ -n "$$gobin" ]; then printf "%s" "$$gobin"; else printf "%s/bin" "$$(go env GOPATH)"; fi')
GOOSE_BIN := $(GO_BIN_DIR)/goose
MIGRATIONS_DIR := db/migrations

ifneq (,$(wildcard .env))
include .env
export DATABASE_URL
endif

.PHONY: goose-install migrate-up migrate-down migrate-status migrate-create check-db-url

goose-install:
	@if [ ! -x "$(GOOSE_BIN)" ]; then echo "Instalando goose $(GOOSE_VERSION)..."; go install github.com/pressly/goose/v3/cmd/goose@$(GOOSE_VERSION); fi

check-db-url:
	@if [ -z "$(DATABASE_URL)" ]; then echo "DATABASE_URL nao definida (configure no .env ou exporte no shell)."; exit 1; fi

migrate-up: goose-install check-db-url
	@$(GOOSE_BIN) -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" up

migrate-down: goose-install check-db-url
	@$(GOOSE_BIN) -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" down

migrate-status: goose-install check-db-url
	@$(GOOSE_BIN) -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" status

migrate-create: goose-install
	@if [ -z "$(name)" ]; then echo "Uso: make migrate-create name=nome_da_migration"; exit 1; fi
	@$(GOOSE_BIN) -dir $(MIGRATIONS_DIR) create $(name) sql
