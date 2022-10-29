VERSION :=$(shell git describe --tags)
LDFLAGS := "-s -w -X main.version=$(VERSION)"
OUT_DIR := dist
CMD := ./cmd/rmfakecloud
BINARY := rmfakecloud
BUILD = go build -ldflags $(LDFLAGS) -o $(@) $(CMD) 
ASSETS = new-ui/dist
GOFILES := $(shell find . -iname '*.go' ! -iname "*_test.go")
GOFILES += $(ASSETS)
UIFILES := $(shell find new-ui/src)
UIFILES += $(shell find new-ui/public)
UIFILES += $(shell find new-ui/types)
UIFILES += new-ui/package.json
TARGETS := $(addprefix $(OUT_DIR)/$(BINARY)-, x64 armv6 armv7 arm64 win64 docker)
YARN	= yarn --cwd new-ui  

.PHONY: all run runui clean test testgo testui

build: $(OUT_DIR)/$(BINARY)-x64

all: $(TARGETS)

$(OUT_DIR)/$(BINARY)-x64:$(GOFILES)
	GOOS=linux $(BUILD)

$(OUT_DIR)/$(BINARY)-armv6:$(GOFILES)
	GOARCH=arm GOARM=6 $(BUILD)

$(OUT_DIR)/$(BINARY)-armv7:$(GOFILES)
	GOARCH=arm GOARM=7 $(BUILD)

$(OUT_DIR)/$(BINARY)-win64:$(GOFILES)
	GOOS=windows $(BUILD)

$(OUT_DIR)/$(BINARY)-arm64:$(GOFILES)
	GOARCH=arm64 $(BUILD)

$(OUT_DIR)/$(BINARY)-docker:$(GOFILES)
	CGO_ENABLED=0 $(BUILD)

container: $(OUT_DIR)/$(BINARY)-docker
	docker build -t rmfakecloud -f Dockerfile.make .
	
run: $(ASSETS)
	go run $(CMD) $(ARG)

$(ASSETS): $(UIFILES) new-ui/yarn.lock
	$(YARN) build

new-ui/yarn.lock: new-ui/node_modules new-ui/package.json
	$(YARN)
	@touch -mr $(shell ls -Atd $? | head -1) $@

new-ui/node_modules:
	mkdir -p $@

runui: new-ui/yarn.lock
	$(YARN) dev

clean:
	rm -f $(OUT_DIR)/*
	rm -fr $(ASSETS)

test: testui testgo

testui:
	CI=true $(YARN) test

testgo:
	go test ./...

