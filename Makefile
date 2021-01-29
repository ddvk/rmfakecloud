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
all: prep $(TARGETS)

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
	yarn --cwd ui run build
	@#remove unneeded stuff, todo: eject
	@rm ui/build/service-worker.js ui/build/precache-manifest* ui/build/asset-manifest.json 2> /dev/null || true

.PHONY: run
run: 
	go run $(CMD)

dev:
	find . -path ui -prune -false -o -iname "*.go" | entr -r go run -tags dev $(CMD)
devui:
	yarn --cwd ui start

#install ui stuff
prep:
	yarn --cwd ui install

.PHONY: clean
clean:
	rm -f $(OUT_DIR)/*

.PHONY: test
test: 
	go test ./...
	CI=true yarn --cwd ui test

