/*
 * Copyright (C) 2024 by Jason Figge
 */

package models

import (
	"us.figge.auto-ssh/internal/core/config"
)

type Tunnel interface {
	Tunnels() []*config.Tunnel
}

//type TunnelEntry struct {
//	Name          string
//	Group         string
//	LocalPort     uint16
//	RemoteAddress string
//	RemotePort    uint16
//	JumpTunnel    string
//	Metadata      *config.Metadata
//}
