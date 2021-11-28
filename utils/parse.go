package utils

import (
	"strings"
)

// ParseArgs parse command line arguments
func ParseArgs(args string) []string {
	s := strings.Split(args, ",")

	return s
}
