.PHONY: api.test api.lint api.fmt web.lint web.fmt all

api.test:
	cd api && go test ./...

api.lint:
	cd api && golangci-lint run ./...

api.fmt:
	cd api && go fmt ./...

web.lint:
	cd web && npm run lint

web.fmt:
	cd web && npm run prettier:check

all: api.fmt api.lint api.test web.fmt web.lint
