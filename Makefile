BIN_DIR := ./bin
INC := 1


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
	@for i in $$(seq 1 $(INC)); do \
		echo -n "Iteration $$i: "; \
		./$< \
			-test.run=^TestIteration$$i[AB]*$$ \
			-binary-path=$(BIN_DIR)/server \
			-agent-binary-path=$(BIN_DIR)/agent; \
	done

.PHONY: clean
clean:
	rm -rf $(BIN_DIR)


$(BIN_DIR)/metricstest: .FORCE
	cd tools/go-autotests && go test -c -o ../../$@ ./cmd/$(@F)

$(BIN_DIR)/server $(BIN_DIR)/agent: $(BIN_DIR)/%: cmd/% .FORCE
	go build -o $@ ./$<
