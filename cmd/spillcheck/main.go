// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

// Command spillcheck is the AgentSpillCheck binary (PLAN D2).
package main

import (
	"os"

	"github.com/Hungleminh/agentspillcheck/internal/cli"
)

// version is set via -ldflags "-X main.version=..." at release time.
var version = "dev"

func main() {
	if err := cli.NewRootCommand(version, os.Stdout).Execute(); err != nil {
		os.Exit(1)
	}
}
