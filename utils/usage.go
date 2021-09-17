package utils

import (
	"flag"
	"fmt"
	"os"
)

const (
	usage = `usage: %s [OPTIONS] <COMMA_SEPPARATED_NODE_NAMES>

Simple tool for safe draining nodes, rolling out deployments without downtime.

Options:
`
)

func Usage() {
	fmt.Fprintf(flag.CommandLine.Output(), usage, os.Args[0])
	flag.PrintDefaults()
}
