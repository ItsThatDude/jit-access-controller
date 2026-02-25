package main

import (
	"github.com/itsthatdude/jit-access-controller/internal/plugin"
)

var Version string

func main() {
	if Version == "" {
		Version = "development"
	}

	plugin.Init()
	plugin.Execute()
}
