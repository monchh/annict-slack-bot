# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
BOT_BINARY=annict-slack-bot
NOTIFIER_BINARY=slack-notifier
BOT_CMD_PATH=./cmd
NOTIFIER_CMD_PATH=./cmd/slack_notifier

# Default target
all: build-all

# Build targets
build: build-bot

build-bot:
	@echo "Building $(BOT_BINARY)..."
	$(GOBUILD) -o $(BOT_BINARY) $(BOT_CMD_PATH)
	@echo "$(BOT_BINARY) built successfully."

build-notifier:
	@echo "Building $(NOTIFIER_BINARY)..."
	$(GOBUILD) -o $(NOTIFIER_BINARY) $(NOTIFIER_CMD_PATH)
	@echo "$(NOTIFIER_BINARY) built successfully."

build-all: build-bot build-notifier

# Run targets
run: run-bot

run-bot: build-bot
	@echo "Running $(BOT_BINARY)..."
	./$(BOT_BINARY)

run-notifier: build-notifier
	@echo "Running $(NOTIFIER_BINARY)..."
	./$(NOTIFIER_BINARY)

# Clean the built binary
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(NOTIFIER_BINARY)
	@echo "Cleaned."

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOGET) ./...
	@echo "Dependencies downloaded."

.PHONY: all build build-bot build-notifier build-all run run-bot run-notifier clean deps

# Generate GraphQL client
make-client:
	@echo "Generating GraphQL client..."
	mkdir -p schema/annict
	curl -o schema/annict/schema.graphql https://raw.githubusercontent.com/annict/annict/main/app/graphql/beta/schema.graphql
	go run github.com/Yamashou/gqlgenc
	@echo "GraphQL client generated."
