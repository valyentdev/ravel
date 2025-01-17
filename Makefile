
run-raveld:
	sudo go run cmd/ravel/ravel.go daemon -c ravel.toml --debug
run-api:
	air
run-corro:
	sudo corrosion agent -c docs/examples/corrosion-config.toml
build-ravel:
	CGO_ENABLED=0 go build -o bin/ravel cmd/ravel/ravel.go
build-initd:
	CGO_ENABLED=0 go build -o bin/initd cmd/initd/initd.go
build-jailer:
	CGO_ENABLED=0 go build -o bin/jailer cmd/jailer/jailer.go

install-ravel: build-ravel
	sudo cp ./bin/ravel /usr/bin/ravel
run-edge:
	sudo go run cmd/ravel-proxy/ravel-proxy.go start -m edge -c proxy.toml --debug
run-local:
	sudo go run cmd/ravel-proxy/ravel-proxy.go start -m local -c proxy.toml --debug