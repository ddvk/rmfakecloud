VERSION :=$(shell git describe --tags)
LDFLAGS := "-s -w -X main.version=$(VERSION)"
OUT_DIR := dist
CMD := ./cmd/rmfakecloud
BINARY := rmfakecloud
BUILD = go build -ldflags $(LDFLAGS) -o $(@) $(CMD) 
ASSETS = ui/build
GOFILES := $(shell find . -iname '*.go' ! -iname "*_test.go")
GOFILES += $(ASSETS)
UIFILES := $(shell find ui/src)
UIFILES += $(shell find ui/public)
UIFILES += ui/package.json
TARGETS := $(addprefix $(OUT_DIR)/$(BINARY)-, x64 armv6 armv7 arm64 win64 docker)
YARN	= yarn --cwd ui  

.PHONY: all run dev devui clean test 
all: $(TARGETS)

build: $(OUT_DIR)/$(BINARY)-x64

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
	go run -tags dev $(CMD)

dev: 
	find . -path ui -prune -false -o -iname "*.go" | entr -r  go run -tags dev $(CMD)

$(ASSETS): $(UIFILES) ui/yarn.lock
	$(YARN) build
	@#remove unneeded stuff, todo: eject
	@rm ui/build/service-worker.js ui/build/precache-manifest* ui/build/asset-manifest.json 2> /dev/null || true

ui/yarn.lock: ui/node_modules ui/package.json
	$(YARN)
	@touch -mr $(shell ls -Atd $? | head -1) $@

ui/node_modules:
	mkdir -p $@

devui: ui/yarn.lock
	$(YARN) start

clean:
	rm -f $(OUT_DIR)/*
	rm -fr $(ASSETS)

test: 
	go test -tags ci ./...
	CI=true $(YARN) test

