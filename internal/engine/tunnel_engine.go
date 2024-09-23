/*
 * Copyright (C) 2024 by Jason Figge
 */

package engine

import (
	"context"

	"us.figge.auto-ssh/internal/core/config"
)

type Tunnel struct {
	Name          string
	Group         string
	LocalPort     uint16
	RemoteAddress string
	RemotePort    uint16
	JumpHost      string
	Metadata      *config.Metadata
}

func NewTunnelEngine(ctx context.Context, cfg config.Tunnel) (*Tunnel, error) {
	engine := &Tunnel{}
	return engine, nil
}
