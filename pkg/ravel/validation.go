package ravel

import "fmt"

func validateObjectName(name string) error {
	if name == "" {
		return fmt.Errorf("fleet name cannot be empty")
	}

	if len(name) > 63 {
		return fmt.Errorf("fleet name cannot be longer than 63 characters")
	}

	for _, c := range name {
		if !(c == '-' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			return fmt.Errorf("fleet name can only contain alphanumeric characters and dashes")
		}
	}

	if name[0] == '-' || name[len(name)-1] == '-' {
		return fmt.Errorf("fleet name cannot start or end with a dash")
	}

	return nil
}
