VERSION := $(shell git describe --tags)
LDFLAGS := "-s -w -X main.version=$(VERSION)"
OUT_DIR := dist
CMD := ./cmd/rmfakecloud
BINARY := rmfakecloud
BUILD = go build -ldflags $(LDFLAGS) -o $(@) $(CMD) 
ASSETS = ui/build
NEWUI_ASSETS = new-ui/build
GOFILES := $(shell find . -iname '*.go' ! -iname "*_test.go")
GOFILES += $(ASSETS)
GOFILES += $(NEWUI_ASSETS)
UIFILES := $(shell find ui/src)
UIFILES += $(shell find ui/public)
UIFILES += ui/package.json
NEWUIFILES := $(shell find new-ui/src)
NEWUIFILES += $(shell find new-ui/public)
NEWUIFILES += new-ui/package.json
TARGETS := $(addprefix $(OUT_DIR)/$(BINARY)-, x64 armv6 armv7 arm64 win64 docker)
YARN = yarn --cwd ui
NEWUI_YARN = yarn --cwd new-ui
HWR_GENERATED_DIR := "./internal/hwr/generated/"
HWR_SWAGGER := "https://cloud.myscript.com/api/v4.0/iink/batch/api-docs"
HWR_PACKAGE := generated

.PHONY: all run runui clean test testgo testui generate_hwr_client run-newui

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

new-run: $(NEWUI_ASSETS)
	go run $(CMD) $(ARG)

$(ASSETS): $(UIFILES) ui/yarn.lock
	@cp ui/node_modules/pdfjs-dist/build/pdf.worker.js ui/public/
	$(YARN) build
	@#remove unneeded stuff, todo: eject
	@rm ui/build/service-worker.js ui/build/precache-manifest* ui/build/asset-manifest.json 2> /dev/null || true

$(NEWUI_ASSETS): $(NEWUIFILES) new-ui/yarn.lock
	$(NEWUI_YARN) build

ui/yarn.lock: ui/node_modules ui/package.json
	$(YARN)
	@touch -mr $(shell ls -Atd $? | head -1) $@

ui/node_modules:
	mkdir -p $@

new-ui/yarn.lock: new-ui/node_modules new-ui/package.json
	$(NEWUI_YARN)
	@touch -mr $(shell ls -Atd $? | head -1) $@

new-ui/node_modules:
	mkdir -p $@

runui: ui/yarn.lock
	$(YARN) start

run-newui: new-ui/yarn.lock
	$(NEWUI_YARN) start

clean:
	rm -rf $(HWR_GENERATED_DIR)
	rm -f $(OUT_DIR)/*
	rm -fr $(ASSETS)
	rm -rf $(NEWUI_ASSETS)

test: testui testgo

testui:
	CI=true $(YARN) test

generate_hwr_client:
	swagger-codegen generate -i $(HWR_SWAGGER) -l go -o $(HWR_GENERATED_DIR) --additional-properties packageName=$(HWR_PACKAGE)

testgo:
	go test ./...

