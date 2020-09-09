VERSION :=$(shell git describe --tags)
LDFLAGS := "-s -w -X main.version=$(VERSION)"
OUT_DIR := dist
CMD := ./cmd/rmfakecloud
BINARY := rmfakecloud
BUILD = go build -ldflags $(LDFLAGS) -o $(@) $(CMD) 
GOFILES := $(shell find . -iname '*.go' ! -iname "*_test.go")
TARGETS := $(addprefix $(OUT_DIR)/$(BINARY)-, x64 armv6 armv7 win64 docker)

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
	docker build -t rmfakecloud .
	

ui:
	npm build --prefix ui

run: 
	go run $(CMD)


clean:
	rm -f $(OUT_DIR)/*

test: 
	go test ./...

