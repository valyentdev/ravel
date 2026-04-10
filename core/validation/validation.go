package validation

func joinFieldPath(previous string, current string) string {
	if previous == "" {
		return current
	}
	return previous + "." + current
}
