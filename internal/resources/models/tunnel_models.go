/*
 * Copyright (C) 2024 by Jason Figge
 */

package models

import (
	"context"
	"sync"

	"us.figge.auto-ssh/internal/core/config"
)

type TunnelEngine interface {
	Tunnels() []Tunnel
	Tunnel(string) (Tunnel, bool)
	StartTunnels(ctx context.Context, stats StatsEngine, wg *sync.WaitGroup)
}

type Tunnel interface {
	Id() string
	Name() string
	Local() *config.Address
	Remote() *config.Address
	Host() string
	Valid() bool
	Running() string
	Metadata() *config.Metadata
	Start()
	Stop()
}
