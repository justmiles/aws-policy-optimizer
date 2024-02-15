package main

import (
	"github.com/justmiles/aws-policy-optimizer/cmd"
)

// version of aws-policy-optimizer. Overwritten during build
var version = "0.0.0"

func main() {
	cmd.Execute(version)
}
