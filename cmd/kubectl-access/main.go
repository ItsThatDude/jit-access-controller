package main

import (
	"antware.xyz/jitaccess/internal/plugin"
)

var Version string

func main() {
	if Version == "" {
		Version = "development"
	}

	plugin.Execute()
}
