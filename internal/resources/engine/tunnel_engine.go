/*
 * Copyright (C) 2024 by Jason Figge
 */

package engine

import (
	"context"

	"us.figge.auto-ssh/internal/core/config"
)

type TunnelEngine struct {
	tunnels []*config.Tunnel
}

func NewTunnelEngine(ctx context.Context, cfg []*config.Tunnel) (*TunnelEngine, error) {
	engine := &TunnelEngine{
		tunnels: cfg,
	}
	return engine, nil
}

func (t TunnelEngine) Tunnels() []*config.Tunnel {
	return t.tunnels
}
