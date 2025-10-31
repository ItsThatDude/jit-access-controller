package main

import (
	"antware.xyz/kairos/internal/plugin"
)

var Version string

func main() {
	if Version == "" {
		Version = "development"
	}

	plugin.Init()
	plugin.Execute()
}
