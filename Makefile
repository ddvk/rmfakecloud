VERSION :=$(shell git describe --tags)
LDFLAGS := "-s -w -X main.version=$(VERSION)"
OUT_DIR := dist
CMD := ./cmd/rmfakecloud
BINARY := rmfakecloud
BUILD = go build -ldflags $(LDFLAGS) -o $(@) $(CMD) 
GOFILES := $(shell find . -iname '*.go' ! -iname "*_test.go")
GOFILES += assets_vfsdata.go
UIFILES := $(shell find ui/src)
UIFILES += $(shell find ui/public)
TARGETS := $(addprefix $(OUT_DIR)/$(BINARY)-, x64 armv6 armv7 win64 docker)

.PHONY: all
all: $(TARGETS)

$(OUT_DIR)/$(BINARY)-x64: $(GOFILES)
	GOOS=linux $(BUILD)

$(OUT_DIR)/$(BINARY)-armv6:$(GOFILES)
	GOARCH=arm GOARM=6 $(BUILD)

$(OUT_DIR)/$(BINARY)-armv7:$(GOFILES)
	GOARCH=arm GOARM=7 $(BUILD)

$(OUT_DIR)/$(BINARY)-win64:$(GOFILES)
	GOOS=windows $(BUILD)

$(OUT_DIR)/$(BINARY)-docker:$(GOFILES)
	CGO_ENABLED=0 $(BUILD)

container: $(OUT_DIR)/$(BINARY)-docker
	docker build -t rmfakecloud -f Dockerfile.make .
	
assets_vfsdata.go: ui
	go generate ./...

.PHONY: ui
ui: $(UIFILES)
	npm run build --prefix ui
	#remove unneeded stuff, todo: eject
	rm ui/build/service-worker.js ui/build/precache-manifest* ui/build/asset-manifest.json|| true

.PHONY: run
run: 
	go run $(CMD)

dev:
	@bash -c 'ls'
	find . -path ui -prune -false -o -iname "*.go" | entr -r go run $(CMD)
devui:
	npm start --prefix ui

prep:
	npm i --prefix ui

.PHONY: clean
clean:
	rm -f $(OUT_DIR)/*

.PHONY: test
test: 
	go test ./...

