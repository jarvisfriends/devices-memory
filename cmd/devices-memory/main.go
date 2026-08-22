// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

// Command devices-memory reports RAM specs and live readings using the
// shared devices-common CLI: one-shot, --json, --every 0.25, --tui, --web.
package main

import (
	"os"

	"github.com/jarvisfriends/devices-common/cli"
	"github.com/jarvisfriends/devices-common/tui"

	memory "github.com/jarvisfriends/devices-memory"
)

// version is stamped by goreleaser via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	os.Exit(cli.Main(cli.Config{
		Collector: memory.New(),
		Version:   version,
		TUI:       tui.Run,
	}, os.Args[1:], os.Stdout, os.Stderr))
}
