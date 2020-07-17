LDFLAGS:="-s -w"
OUT_DIR:=dist

run:
	go run .

all: 
	go build -o $(OUT_DIR)/rmfakecloud-x64
	GOARCH=arm GOARM=7 go build -ldflags $(LDFLAGS) -o $(OUT_DIR)/rmfakecloud-armv7
	GOARCH=arm GOARM=6 go build -ldflags $(LDFLAGS) -o $(OUT_DIR)/rmfakecloud-armv6
	GOOS=windows go build -ldflags $(LDFLAGS) -o $(OUT_DIR)/rmfakecloud-winx64

$(OUT_DIR)/rmfake-docker:
	CGO_ENABLED=0 go build -o $@

docker: $(OUT_DIR)/rmfake-docker
	docker build -t rmfakecloud .
	
clean:
	rm -f $(OUT_DIR)/*

test:
	@echo "some test"



