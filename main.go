package main

import (
	"github.com/shadowbq/simple-node-health/cmd"
)

var (
	Version string // set by the build process
)

func main() {
	cmd.Execute()
}
