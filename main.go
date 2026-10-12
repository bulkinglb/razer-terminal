package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/bulkinglb/razer-terminal/internal/razer"
	"github.com/bulkinglb/razer-terminal/internal/tui"
)

const usage = `usage: razer-terminal

Opens a settings editor for Razer devices driven by OpenRazer.

`

// version is set by the Makefile for release builds.
var version = "dev"

func main() {
	showVersion := flag.Bool("version", false, "print the version")
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
		flag.PrintDefaults()
	}
	flag.Parse()
	if *showVersion {
		// go install records the module version instead of the Makefile's.
		if info, ok := debug.ReadBuildInfo(); ok && version == "dev" && info.Main.Version != "" && info.Main.Version != "(devel)" {
			version = info.Main.Version
		}
		fmt.Println("razer-terminal", version)
		return
	}

	devs, err := razer.Devices()
	if len(devs) == 0 {
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
		}
		fmt.Fprintln(os.Stderr, `no Razer devices found. Check that the openrazer driver is installed, the
daemon runs (systemctl --user status openrazer-daemon) and you're in the
plugdev group.`)
		os.Exit(1)
	}
	status := ""
	if err != nil {
		status = "MICE AND KEYBOARDS SKIPPED: " + err.Error()
	}
	if err := tui.Run(devs, status); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
