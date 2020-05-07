LDFLAGS:="-s -w"
OUT:=out

run:
	go run .

all: 
	go build -o $(OUT)/rmfakecloud-x64
	GOARCH=arm GOARM=7 go build -ldflags $(LDFLAGS) -o $(OUT)/rmfakecloud-armv7
	GOARCH=arm GOARM=6 go build -ldflags $(LDFLAGS) -o $(OUT)/rmfakecloud-armv6
	GOOS=windows go build -ldflags $(LDFLAGS) -o $(OUT)/rmfakecloud-winx64

$(OUT)/rmfake-docker:
	CGO_ENABLED=0 go build -o $@

docker: $(OUT)/rmfake-docker
	docker build -t rmfakecloud .
	
clean:
	rm -f $(OUT)/*

test:
	@echo "some test"



