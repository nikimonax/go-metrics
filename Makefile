BIN_DIR := ./bin

ITER ?= $(shell \
	git branch --show-current | \
	sed -n 's/^iter\([0-9]\+\)$$/\1/p')


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

.PHONY: autotest
autotest: $(BIN_DIR)/metricstest $(BIN_DIR)/server $(BIN_DIR)/agent
	@if [ -z "$(ITER)" ]; then \
		echo "\nTest iteration could not be determined."; \
		echo "Please provide 'ITER' variable.\n"; \
		exit 1; \
	fi; \
	for i in $$(seq 1 $(ITER)); do \
		echo -n "Iteration $$i: "; \
		./$< \
			-test.run=^TestIteration$$i[AB]*$$ \
			-binary-path=$(BIN_DIR)/server \
			-agent-binary-path=$(BIN_DIR)/agent \
			-source-path=.; \
	done

.PHONY: clean
clean:
	rm -rf $(BIN_DIR)


$(BIN_DIR)/metricstest: .FORCE
	cd tools/go-autotests && go test -c -o ../../$@ ./cmd/$(@F)

$(BIN_DIR)/server $(BIN_DIR)/agent: $(BIN_DIR)/%: cmd/% .FORCE
	go build -o $@ ./$<
