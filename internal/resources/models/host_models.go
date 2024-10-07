/*
 * Copyright (C) 2024 by Jason Figge
 */

package models

import (
	"us.figge.auto-ssh/internal/core/config"
)

type HostEngine interface {
	Hosts() []Host
	Host(string) (Host, bool)
	KnownHosts() []string
	MarkInUse(name string)
}

type Host interface {
	Id() string
	Name() string
	Remote() *config.Address
	Username() string
	Identity() string
	KnownHosts() string
	JumpHost() string
	Valid() bool
	Metadata() *config.Metadata
}
