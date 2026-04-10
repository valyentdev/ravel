package cloudhypervisor

func StringPtr(s string) *string {
	if s == "" {
		return nil
	}

	return &s
}

func BoolPtr(b bool) *bool {
	return &b
}
