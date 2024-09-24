/*
 * Copyright (C) 2024 by Jason Figge
 */

package models

import (
	"us.figge.auto-ssh/internal/core/config"
)

type Host interface {
}

type HostEntry struct {
	Name       string
	Group      string
	Address    string
	Username   string
	Identity   string
	KnownHosts string
	JumpHost   string
	Metadata   *config.Metadata
}
