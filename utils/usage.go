/*
 * Copyright (c) 2021 Angel Abad. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

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

// Usage shows flag usage help
func Usage() {
	fmt.Fprintf(flag.CommandLine.Output(), usage, os.Args[0])
	flag.PrintDefaults()
}
