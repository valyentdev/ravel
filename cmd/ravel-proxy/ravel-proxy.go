package main

import (
	"fmt"
	"os"

	"github.com/valyentdev/ravel/cmd/ravel-proxy/commands"
)

func main() {
	if err := commands.NewRootCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
