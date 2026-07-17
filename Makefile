BIN_DIR := ./bin

COV_FILE := coverage.out
COV_FILE_HTML := coverage.html

TEST_ARGS += $(ARGS)


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
lint:
	golangci-lint run

.PHONY: test
test:
	go test $(TEST_ARGS) $$(go list ./... | grep -v internal/mock)

.PHONY: cover
cover: $(COV_FILE)
	go tool cover -func=$(COV_FILE)

.PHONY: cover-html
cover-html: $(COV_FILE)
	go tool cover -html=$(COV_FILE) -o $(COV_FILE_HTML)

.PHONY: autotest
autotest: $(BIN_DIR)/metricstest $(BIN_DIR)/server $(BIN_DIR)/agent
	./tools/autotest.sh

.PHONY: clean
clean:
	rm -rf $(BIN_DIR)
	go clean -testcache

$(COV_FILE): TEST_ARGS += -coverprofile=$(COV_FILE)
$(COV_FILE): test

$(BIN_DIR)/metricstest: .FORCE
	cd tools/go-autotests && go test -c -o ../../$@ ./cmd/$(@F)

$(BIN_DIR)/server $(BIN_DIR)/agent: $(BIN_DIR)/%: cmd/% .FORCE
	go build -o $@ ./$<
