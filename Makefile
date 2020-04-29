LDFLAGS:="-ldflags -s -w"

run:
	go run .

build:
	go build -o bin/rmfakecloud-x64
	GOARCH=arm GOARM=7 go build $(LDFLAGS) -o bin/rmfakecloud-armv7
	GOARCH=arm GOARM=6 go build $(LDFLAGS) -o bin/rmfakecloud-armv6
test:
	@echo "some test"



