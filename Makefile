VERSION := $(shell date +%Y.%m.%d.%H%M%S)
LDFLAGS := "-s -w -X main.version=$(VERSION)"
OUT_DIR := dist
CMD := ./cmd/rmfakecloud
BINARY := rmfakecloud
BUILD = go build -ldflags $(LDFLAGS) -o $(@) $(CMD)
GOFILES := $(shell find . -iname '*.go' ! -iname "*_test.go")
STATICFILES := $(shell find internal/ui/static -type f 2>/dev/null)
THEMEFILES := $(shell find internal/ui/themes -type f 2>/dev/null)
GOFILES += $(STATICFILES)
GOFILES += $(THEMEFILES)
TARGETS := $(addprefix $(OUT_DIR)/$(BINARY)-, x64 armv6 armv7 arm64 win64 docker)

.PHONY: all run clean test testgo build

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

run:
	go run $(CMD) $(ARG)

clean:
	rm -f $(OUT_DIR)/*

test: testgo

testgo:
	go test ./...
