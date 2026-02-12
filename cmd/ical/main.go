package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/BRO3886/ical/cmd/ical/commands"
)

// Set by ldflags at build time.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if runtime.GOOS != "darwin" {
		fmt.Fprintln(os.Stderr, "ical requires macOS")
		os.Exit(1)
	}

	commands.SetVersionInfo(version, commit, date)
	if err := commands.Execute(); err != nil {
		os.Exit(1)
	}
}
