package main

import (
	"fmt"
	"os"

	"github.com/valyentdev/ravel/cmd/ravel/commands"
)

func main() {
	rootCmd := commands.NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
