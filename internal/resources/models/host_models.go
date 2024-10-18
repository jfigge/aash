/*
 * Copyright (C) 2024 by Jason Figge
 */

package models

import (
	"net"

	"us.figge.auto-ssh/internal/core/config"
	engineTunnel "us.figge.auto-ssh/internal/resources/engine/tunnel"
)

type HostEngine interface {
	Hosts() []Host
	Host(string) (Host, bool)
	KnownHosts() []string
}

type HostEngineInternal interface {
	HostEngine
	ValidateJumpHosts(tunnelEntries map[string]*engineTunnel.Entry)
}

type Host interface {
	Id() string
	Name() string
	Remote() *config.Address
	Username() string
	Passphrase() string
	Identity() string
	KnownHosts() string
	JumpHost() string
	Valid() bool
	Metadata() *config.Metadata
}

type HostInternal interface {
	Host
	Open() bool
	Dial(address string) (net.Conn, bool)
	Referenced()
}
