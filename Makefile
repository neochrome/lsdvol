.PHONY: build clean
.DEFAULT_GOAL := build

VERSION=$(shell git describe --tags --dirty 2>/dev/null || echo 'dev')
BINARY=bin/lsdvol

build: $(BINARY)

$(BINARY): *.go
	@echo building version: $(VERSION)
	@mkdir -p bin
	@CGO_ENABLED=0 go build -o $(BINARY) -ldflags "-X main.version=$(VERSION)"

clean:
	@echo cleaning
	@rm -rf bin
