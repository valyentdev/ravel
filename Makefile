build-init:
	CGO_ENABLED=0 go build -o bin/ravel-init -ldflags="-s -w" cmd/ravel-init/*.go 

build-ravel: build-init protoc
	go build -o ./bin/ravel cmd/ravel/main.go

install-ravel: build-ravel
	sudo cp ./bin/ravel /usr/local/bin/ravel


protoc:
	buf generate
