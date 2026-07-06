.PHONY: build test lint check

build:
	cd app/frontend && npm run build
	cd app && GOCACHE=/private/tmp/thanos-go-build GOMODCACHE=/private/tmp/thanos-go-mod go build ./...

test:
	go test ./...
	cd app && GOCACHE=/private/tmp/thanos-go-build GOMODCACHE=/private/tmp/thanos-go-mod go test ./...
	cd app/frontend && npm run test

lint:
	go vet ./...

check: build test lint
