# Thanos — build & run
#
# Two parts:
#   backend/  Go daemon + `to` CLI (Thanos Orchestrator)
#   app/      Electron + Vite + React desktop app
#
# Override the package manager with `make PM=npm <target>` if you don't use pnpm.

PM  ?= pnpm
GO  ?= go
BIN := bin
# Runs a command with a Node >=22.12 on PATH (Vite 8 requirement); auto-selects an
# nvm install when the system Node is too old. See scripts/with-node.sh.
NODE := ./scripts/with-node.sh

.DEFAULT_GOAL := help

## ----------------------------------------------------------------------------
## Run
## ----------------------------------------------------------------------------

.PHONY: run dev
run: dev ## Alias for `dev`: launch the desktop app
dev: app/node_modules ## Run the desktop app (Electron); spawns the daemon via `go run ./cmd/to daemon`
	cd app && $(CURDIR)/$(NODE) $(PM) run dev

.PHONY: dev-web
dev-web: app/node_modules ## Run only the renderer in a browser (no Electron); needs a daemon running
	cd app && $(CURDIR)/$(NODE) $(PM) run dev:web

.PHONY: daemon
daemon: ## Run the backend daemon directly in the foreground
	cd backend && $(GO) run ./cmd/to daemon

## ----------------------------------------------------------------------------
## Build
## ----------------------------------------------------------------------------

.PHONY: build
build: build-backend build-app ## Build the `to` CLI and package the desktop app

.PHONY: build-backend cli
build-backend cli: ## Build the `to` CLI binary into ./bin/to
	mkdir -p $(BIN)
	cd backend && $(GO) build -o ../$(BIN)/to ./cmd/to
	@echo "built $(BIN)/to"

.PHONY: build-daemon
build-daemon: ## Build the bundled daemon binary into app/daemon/
	cd app && $(CURDIR)/$(NODE) $(PM) run build:daemon

.PHONY: build-app package
build-app package: app/node_modules ## Package the desktop app (electron-forge package)
	cd app && $(CURDIR)/$(NODE) $(PM) run package

.PHONY: dist make-app
dist make-app: app/node_modules ## Build distributable installers (electron-forge make)
	cd app && $(CURDIR)/$(NODE) $(PM) run make

## ----------------------------------------------------------------------------
## Codegen, tests, lint
## ----------------------------------------------------------------------------

.PHONY: api
api: ## Regenerate the OpenAPI spec and the app's TS API types
	$(PM) run api

.PHONY: sqlc
sqlc: ## Regenerate backend sqlc code from queries/migrations
	$(PM) run sqlc

.PHONY: icons
icons: ## Regenerate app icons from thanos-logo.svg
	./scripts/generate-icons.sh

.PHONY: test test-backend test-app
test: test-backend test-app ## Run all tests
test-backend: ## Run backend Go tests
	cd backend && $(GO) test ./...
test-app: app/node_modules ## Run app unit tests (vitest)
	cd app && $(CURDIR)/$(NODE) $(PM) run test

.PHONY: typecheck
typecheck: app/node_modules ## TypeScript typecheck the app
	cd app && $(CURDIR)/$(NODE) $(PM) run typecheck

.PHONY: lint
lint: ## Backend go test + golangci-lint
	$(PM) run lint

## ----------------------------------------------------------------------------
## Setup & housekeeping
## ----------------------------------------------------------------------------

.PHONY: install
install: app/node_modules ## Install app dependencies

app/node_modules: app/package.json
	cd app && $(CURDIR)/$(NODE) $(PM) install
	@touch app/node_modules

.PHONY: clean
clean: ## Remove build artifacts (bin/, app/daemon/, app/.vite, app/out)
	rm -rf $(BIN) app/daemon app/.vite app/out

.PHONY: help
help: ## Show this help
	@grep -hE '^[a-zA-Z_-]+( [a-zA-Z_-]+)*:.*?## ' $(MAKEFILE_LIST) \
		| sort \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'
