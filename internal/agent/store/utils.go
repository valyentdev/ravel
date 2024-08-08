package store

import (
	"fmt"
	"strings"
)

type scannable interface {
	Scan(dest ...interface{}) error
}

func placeholders(n int) string {
	holders := make([]string, n)
	for i := 0; i < n; i++ {
		holders[i] = fmt.Sprintf("?%d", i+1)
	}
	return strings.Join(holders, ", ")
}

func allColumns(fields []string) string {
	return "`" + strings.Join(fields, "`, `") + "`"
}
