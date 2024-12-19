
run-raveld:
	sudo go run cmd/ravel/ravel.go daemon -c ravel.toml --debug
run-api:
	air
build-ravel:
	CGO_ENABLED=0 go build -o bin/ravel cmd/ravel/ravel.go

build-jailer:
	CGO_ENABLED=0 go build -o bin/jailer cmd/jailer/jailer.go

install-ravel: build-ravel
	sudo cp ./bin/ravel /usr/bin/ravel
protoc:
	buf generate
