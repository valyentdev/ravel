package jailer

import (
	"fmt"
	"os"
	"slices"
)

func validateConfig(config *JailerConfig) error {
	if config.NewRoot == "" {
		return fmt.Errorf("missing new root directory")
	}
	if len(config.Command) == 0 {
		return fmt.Errorf("missing command")
	}
	if config.Fsize < 0 {
		return fmt.Errorf("invalid file size limit")
	}

	if config.NoFiles < 0 {
		return fmt.Errorf("invalid number of open files")
	}

	return nil
}

func parseArgs() (*JailerConfig, error) {
	args := os.Args[2:]

	sepIdx := slices.Index(args, "--")
	if sepIdx == -1 {
		return nil, fmt.Errorf("missing '--' separator and command")
	}
	if sepIdx+1 >= len(args) {
		return nil, fmt.Errorf("missing command")
	}

	jailerArgs := args[:sepIdx]
	command := args[sepIdx+1:]

	var config JailerConfig
	jailerFlagSet := setupJailerFlags(&config)
	err := jailerFlagSet.Parse(jailerArgs)
	if err != nil {
		return nil, err
	}

	config.Command = command

	return &config, validateConfig(&config)
}

func Run() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("usage: %s <run | exec> <args> -- <command args>", os.Args[0])
	}

	action := os.Args[1]

	config, err := parseArgs()
	if err != nil {
		return err
	}

	if action == "exec" {
		return execJailed(config)
	}

	if action == "run" {
		return runJailed(config)
	}

	return fmt.Errorf("unknown action: %s", action)
}
