include .env

BIN_DIR  := ./bin
OUT_DIR  := ./out
DOCS_DIR := ./docs

PPROF_PORT := 8081

COV_FILE := $(OUT_DIR)/coverage.out
COV_FILE_HTML := $(OUT_DIR)/coverage.html

POSTGRES_DSN := "postgres://$(DATABASE_USER):$(DATABASE_PASSWORD)@$(DATABASE_HOST):$(DATABASE_PORT)/$(DATABASE_NAME)?sslmode=disable"
MIGRATIONS_DIR := internal/impl/repository/migrations

EXTRA_ARGS += $(ARGS)


all: build


.SUFFIXES:

.PHONY: .FORCE
.FORCE:

.PHONY: run-server run-agent
run-server run-agent: run-%: $(BIN_DIR)/%
	./$< $(EXTRA_ARGS)

.PHONY: gen generate
gen generate:
	go generate ./...

.PHONY: build
build: $(BIN_DIR)/server $(BIN_DIR)/agent

.PHONY: lint
lint: _check_golangci_lint_cmd
	golangci-lint run

.PHONY: fmt format
fmt format: _check_golangci_lint_cmd _check_swag_cmd
	golangci-lint fmt
	swag fmt

.PHONY: test
test: | $(OUT_DIR)
	go test $(EXTRA_ARGS) $$(go list ./... | grep -v internal/testing)

.PHONY: bench
bench: bench-v1 bench-v2 bench-v3

.PHONY: bench-v1 bench-v2 bench-v3
bench-v1 bench-v2 bench-v3: bench-%: $(BIN_DIR)/bench | $(OUT_DIR)
	$< \
		-test.bench=BenchmarkAPI/$* \
		-test.benchtime=500x \
		-test.benchmem \
		-test.count=3 \
		-test.cpuprofile=$(OUT_DIR)/cpu-$*.profile \
		-test.memprofile=$(OUT_DIR)/mem-$*.profile

pprof-mem-v1 pprof-cpu-v1 \
pprof-mem-v2 pprof-cpu-v2 \
pprof-mem-v3 pprof-cpu-v3: pprof-%: $(OUT_DIR)/%.profile
	go tool pprof \
		-http=:$(PPROF_PORT) \
		-no_browser \
		-ignore='gateway|httptest' \
		$(EXTRA_ARGS) $(BIN_DIR)/bench $<

.PHONY: cover
cover: $(COV_FILE)
	sed -i '/internal\/testing/d' $(COV_FILE) && \
	go tool cover -func=$(COV_FILE)

.PHONY: cover-html
cover-html: $(COV_FILE_HTML)

.PHONY: openapi
swagger: $(DOCS_DIR)/swagger.yaml $(DOCS_DIR)/swagger.json

.PHONY: autotest
autotest: $(BIN_DIR)/metricstest_v2 $(BIN_DIR)/server $(BIN_DIR)/agent
	./tools/autotest.sh $(EXTRA_ARGS)

.PHONY: statictest
statictest: $(BIN_DIR)/statictest
	go vet -vettool=$< ./...

.PHONY: clean
clean:
	rm -rf $(BIN_DIR) $(OUT_DIR)
	go clean -testcache

.PHONY: up down logs
up: EXTRA_ARGS+=-d
up down logs:
	docker compose $@ $(EXTRA_ARGS)

.PHONY: migrate
migrate: _check_migrate_cmd
	migrate -database $(POSTGRES_DSN) -path $(MIGRATIONS_DIR) up


$(COV_FILE): EXTRA_ARGS += -coverpkg=./... -coverprofile=$(COV_FILE)
$(COV_FILE): test | $(OUT_DIR)

$(COV_FILE_HTML): $(COV_FILE) | $(OUT_DIR)
	go tool cover -html=$< -o $@

$(DOCS_DIR)/swagger.yaml $(DOCS_DIR)/swagger.json: | $(DOCS_DIR)
	swag init -d cmd/server,internal/server/handler --pd -o $(DOCS_DIR) -ot json,yaml

$(BIN_DIR)/metricstest_v2: .FORCE
	cd tools/go-autotests && go test -c -o ../../$@ ./cmd/$(@F)

$(BIN_DIR)/statictest: .FORCE
	cd tools/go-autotests && go build -o ../../$@ ./cmd/$(@F)

$(BIN_DIR)/server $(BIN_DIR)/agent: $(BIN_DIR)/%: cmd/% .FORCE
	go build -o $@ ./$<

$(BIN_DIR)/bench: tests .FORCE
	go test -c -o $@ ./$<

$(BIN_DIR) $(OUT_DIR) $(DOCS_DIR):
	mkdir -p $@


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

.PHONY: _check_swag_cmd
_check_swag_cmd:
	@if ! command -v swag $&> /dev/null; then \
		echo "[ERROR]: command 'swag' not found"; \
		echo "install using \"go install github.com/swaggo/swag/cmd/swag@latest"; \
		exit 1; \
	fi
