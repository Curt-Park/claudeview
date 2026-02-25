package main

import "github.com/Curt-Park/claudeview/cmd"

// Version is set at build time via -ldflags "-X main.Version=..."
var Version = "dev"

func main() {
	cmd.AppVersion = Version
	cmd.Execute()
}
