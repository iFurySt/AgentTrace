PROJECT ?=
SLUG ?=
DATABASE_DSN ?= data/agenttrace.db

.PHONY: init new-history new-plan test build serve migrate docker-build

init:
	@if [ -z "$(PROJECT)" ]; then echo "usage: make init PROJECT=my-project"; exit 1; fi
	./scripts/init-project.sh "$(PROJECT)"

new-history:
	@if [ -z "$(SLUG)" ]; then echo "usage: make new-history SLUG=my-change"; exit 1; fi
	./scripts/new-history.sh "$(SLUG)"

new-plan:
	@if [ -z "$(SLUG)" ]; then echo "usage: make new-plan SLUG=my-plan"; exit 1; fi
	./scripts/new-exec-plan.sh "$(SLUG)"

test:
	go test ./...

build:
	go build -o bin/agenttrace ./cmd/agenttrace

serve:
	go run ./cmd/agenttrace serve --database-driver sqlite --database-dsn "$(DATABASE_DSN)"

migrate:
	go run ./cmd/agenttrace migrate --database-driver sqlite --database-dsn "$(DATABASE_DSN)"

docker-build:
	docker build -t agenttrace:local .
