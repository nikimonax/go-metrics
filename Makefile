include .env

BIN_DIR := ./bin

COV_FILE := coverage.out
COV_FILE_HTML := coverage.html

POSTGRES_DSN := "postgres://$(DATABASE_USER):$(DATABASE_PASSWORD)@$(DATABASE_HOST):$(DATABASE_PORT)/$(DATABASE_NAME)?sslmode=disable"
MIGRATIONS_DIR := internal/impl/repository/migrations

EXTRA_ARGS += $(ARGS)


all: build


.SUFFIXES:

.PHONY: .FORCE
.FORCE:

.PHONY: run-server run-agent
run-server run-agent: run-%: $(BIN_DIR)/%
	./$<

.PHONY: build
build: $(BIN_DIR)/server $(BIN_DIR)/agent

.PHONY: lint
lint: _check_golangci_lint_cmd
	golangci-lint run

.PHONY: fmt format
fmt format: _check_golangci_lint_cmd
	golangci-lint fmt

.PHONY: test
test:
	go test $(EXTRA_ARGS) $$(go list ./... | grep -v internal/testing)

.PHONY: cover
cover: $(COV_FILE)
	go tool cover -func=$(COV_FILE)

.PHONY: cover-html
cover-html: $(COV_FILE)
	go tool cover -html=$(COV_FILE) -o $(COV_FILE_HTML)

.PHONY: autotest
autotest: $(BIN_DIR)/metricstest_v2 $(BIN_DIR)/server $(BIN_DIR)/agent
	./tools/autotest.sh $(EXTRA_ARGS)

.PHONY: statictest
statictest: $(BIN_DIR)/statictest
	go vet -vettool=$< ./...

.PHONY: clean
clean:
	rm -rf $(BIN_DIR)
	go clean -testcache

.PHONY: up down logs
up: EXTRA_ARGS+=-d
up down logs:
	docker compose $@ $(EXTRA_ARGS)

.PHONY: migrate
migrate: _check_migrate_cmd
	migrate -database $(POSTGRES_DSN) -path $(MIGRATIONS_DIR) up

$(COV_FILE): EXTRA_ARGS += -coverprofile=$(COV_FILE)
$(COV_FILE): test

$(BIN_DIR)/metricstest_v2: .FORCE
	cd tools/go-autotests && go test -c -o ../../$@ ./cmd/$(@F)

$(BIN_DIR)/statictest: .FORCE
	cd tools/go-autotests && go build -o ../../$@ ./cmd/$(@F)

$(BIN_DIR)/server $(BIN_DIR)/agent: $(BIN_DIR)/%: cmd/% .FORCE
	go build -o $@ ./$<

.env: .env.example
	cp $< $@


.PHONY: _check_golangci_lint_cmd
_check_golangci_lint_cmd:
	@if ! command -v golangci-lint $&> /dev/null; then \
		echo "[ERROR]: command 'golangci-lint' not found"; \
		echo "install using \"go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest\""; \
		exit 1; \
	fi

.PHONY: _check_migrate_cmd 
_check_migrate_cmd:
	@if ! command -v migrate $&> /dev/null; then \
		echo "[ERROR]: command 'migrate' not found"; \
		echo "install using \"go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest\""; \
		exit 1; \
	fi
