/*
 * Copyright (C) 2024 by Jason Figge
 */

package main

import (
	"us.figge.auto-ssh/internal/cmd"
	_ "us.figge.auto-ssh/internal/cmd/core"
	_ "us.figge.auto-ssh/internal/cmd/hosts"
	_ "us.figge.auto-ssh/internal/cmd/tunnels"
)

func main() {
	cmd.Execute()
}
