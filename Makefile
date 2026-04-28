.PHONY: build clean dist

PLUGIN_NAME := oido-rustfs
BINARY := $(PLUGIN_NAME)-mcp
DIST_DIR := dist

build:
	@echo "Building $(PLUGIN_NAME) MCP server..."
	CGO_ENABLED=0 go build -o $(BINARY) .
	@echo "✓ Built: $(BINARY)"
	@ls -lh $(BINARY)

dist: build
	@mkdir -p $(DIST_DIR)
	@echo "Packaging $(PLUGIN_NAME).zip..."
	@cd $(DIST_DIR) && zip -j ../$(DIST_DIR)/$(PLUGIN_NAME).zip \
		../oido-extension.json \
		../OIDO.md \
		../$(BINARY)
	@echo "✓ Packaged: $(DIST_DIR)/$(PLUGIN_NAME).zip"
	@ls -lh $(DIST_DIR)/$(PLUGIN_NAME).zip

clean:
	rm -f $(BINARY)
	rm -rf $(DIST_DIR)
