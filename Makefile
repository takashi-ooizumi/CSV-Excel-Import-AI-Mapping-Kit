.PHONY: fmt test check all api.fmt api.test

fmt: api.fmt
	cd web && npm run prettier:write

api.fmt:
	cd api && goimports -w . && go fmt ./...

test:
	cd api && go vet ./... && go test ./...

check:
	cd web && npm run prettier:check && npm run typecheck
	cd api && go vet ./... && go test ./...

all: fmt check
