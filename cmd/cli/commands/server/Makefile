# Makefile for Go gqlgen server

# Variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOMOD=$(GOCMD) mod
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GQLGEN=gqlgen

# Directories
BINDIR=bin
SRCDIR=.

# Binary name
BINARY_NAME=myapp

# Default target
all: build

# Build the project
build:
	$(GOBUILD) -o $(BINDIR)/$(BINARY_NAME) $(SRCDIR)

# Clean the project
clean:
	$(GOCLEAN)
	rm -f $(BINDIR)/$(BINARY_NAME)

# Run tests
test:
	$(GOTEST) -v ./...

# Run gqlgen to generate GraphQL server code
generate:
	go run github.com/99designs/gqlgen generate

# Tidy up the Go module
tidy:
	$(GOMOD) tidy

# Get dependencies
deps:
	$(GOGET) -u ./...

# Run the application
run: build
	./$(BINDIR)/$(BINARY_NAME)

# Install gqlgen
install-gqlgen:
	$(GOGET) github.com/99designs/gqlgen

# Help message
help:
	@echo "Makefile for Go gqlgen server"
	@echo ""
	@echo "Usage:"
	@echo "  make <target>"
	@echo ""
	@echo "Targets:"
	@echo "  all             Build the project"
	@echo "  build           Build the project"
	@echo "  clean           Clean the project"
	@echo "  test            Run tests"
	@echo "  generate        Run gqlgen to generate GraphQL server code"
	@echo "  tidy            Tidy up the Go module"
	@echo "  deps            Get dependencies"
	@echo "  run             Run the application"
	@echo "  install-gqlgen  Install gqlgen"
	@echo "  help            Show this help message"
