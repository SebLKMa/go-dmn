BIN_DIR := bin

EXAMPLE_DIRS := $(patsubst %/main.go,%,$(wildcard examples/*/main.go))

.PHONY: build clean

build:
	@mkdir -p $(BIN_DIR)
	@for dir in $(EXAMPLE_DIRS); do \
		name=$$(basename $$dir); \
		echo "building $$name"; \
		go build -o $(BIN_DIR)/$$name ./$$dir || exit 1; \
		for f in $$dir/*.json; do \
			[ -e "$$f" ] || continue; \
			echo "copying $$f"; \
			cp "$$f" $(BIN_DIR)/; \
		done; \
	done

clean:
	rm -rf $(BIN_DIR)
