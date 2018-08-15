// Copyright © 2018 ThreeComma.io <hello@threecomma.io>

package main

import "github.com/threecommaio/qflow/cmd"

var (
	// version is set during build
	version = "0.0.1"
)

func main() {
	cmd.Execute(version)
}
