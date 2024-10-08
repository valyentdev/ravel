build-init:
	CGO_ENABLED=0 go build -o bin/ravel-init -ldflags="-s -w" cmd/ravel-init/*.go 

build-ravel:
	go build -o bin/ravel cmd/ravel/ravel.go

install-ravel: build-ravel
	sudo cp ./bin/ravel /usr/bin/ravel
protoc:
	buf generate
