package utils

import "strings"

func ParseArgs(args string) []string {
	s := strings.Split(args, ",")

	return s
}
