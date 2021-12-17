Install and build the project, under Linux:  
`git clone https://github.com/ddvk/rmfakecloud`  
`make`

`make` will just build the x64 binary. If you need binaries for other architectures, you should instead run `make all`.

run  
`~/dist/rmfakecloud-x64`

or clone an do: `go run ./cmd/rmfakecloud`  
or `make run`  
or `make all` artifacts are in the `dist` folder. the Arm binaries work on pi3 / Synology etc  
or `make docker && ./rundocker.sh`
