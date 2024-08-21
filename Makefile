build-init:
	CGO_ENABLED=0 go build -o bin/ravel-init -ldflags="-s -w" cmd/ravel-init/*.go 

build-cli:
	go build -o bin/ravel cmd/ravel/ravel.go
build-raveld:
	go build -o bin/raveld cmd/raveld/raveld.go

install-ravel: build-raveld build-cli
	sudo cp ./bin/ravel /usr/bin/ravel
	sudo cp ./bin/raveld /usr/bin/raveld
protoc:
	buf generate
