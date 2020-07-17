LDFLAGS := "-s -w"
OUT_DIR := dist
BIN ?= ./cmd/rmfakecloud

run:
	go run $(BIN)

all: 
	go build -o $(OUT_DIR)/rmfakecloud-x64 $(BIN) 
	GOARCH=arm GOARM=7 go build -ldflags $(LDFLAGS) -o $(OUT_DIR)/rmfakecloud-armv7 $(BIN)
	GOARCH=arm GOARM=6 go build -ldflags $(LDFLAGS) -o $(OUT_DIR)/rmfakecloud-armv6 $(BIN)
	GOOS=windows go build -ldflags $(LDFLAGS) -o $(OUT_DIR)/rmfakecloud-winx64 $(BIN) 

$(OUT_DIR)/rmfake-docker:
	CGO_ENABLED=0 go build -o $@ $(BIN) 

docker: $(OUT_DIR)/rmfake-docker
	docker build -t rmfakecloud .
	
clean:
	rm -f $(OUT_DIR)/*

test:
	@echo "some test"



