# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
BINARY_NAME=annict-slack-bot
CMD_PATH=./cmd/annict-slack-bot

# Default target
all: build

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) -o $(BINARY_NAME) $(CMD_PATH)
	@echo "$(BINARY_NAME) built successfully."

# Run the application
# Assumes .env file exists or environment variables are set
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_NAME)

# Clean the built binary
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	@echo "Cleaned."

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOGET) ./...
	@echo "Dependencies downloaded."

.PHONY: all build run clean deps

# Generate GraphQL client
make-client:
	@echo "Generating GraphQL client..."
	mkdir -p schema/annict
	curl -o schema/annict/schema.graphql https://raw.githubusercontent.com/annict/annict/main/app/graphql/beta/schema.graphql
	go run github.com/Yamashou/gqlgenc
	@echo "GraphQL client generated."

