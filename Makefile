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
	-@docker rmi lsdvol:test

test: build
	@docker build -t lsdvol:test -f Dockerfile.test .
	@docker run --rm -it \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v $$PWD:/it-works lsdvol:test \
		| grep '^\/it-works'
