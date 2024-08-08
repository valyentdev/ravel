package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

func PopulateProcessEnv(env []string) error {
	for _, pair := range env {
		p := strings.SplitN(pair, "=", 2)
		if len(p) < 2 {
			return errors.New("invalid env var: missing '='")
		}
		name, val := p[0], p[1]
		if name == "" {
			return errors.New("invalid env var: name cannot be empty")
		}
		if strings.IndexByte(name, 0) >= 0 {
			return errors.New("invalid env var: name contains null byte")
		}
		if strings.IndexByte(val, 0) >= 0 {
			return errors.New("invalid env var: value contains null byte")
		}
		if err := os.Setenv(name, val); err != nil {
			return fmt.Errorf("could not set env var: system shit: %v", err)
		}
	}
	return nil
}
