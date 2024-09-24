/*
 * Copyright (C) 2024 by Jason Figge
 */

package engine

import (
	"context"

	"us.figge.auto-ssh/internal/core/config"
	"us.figge.auto-ssh/internal/resources/models"
)

type TunnelEngine struct {
	tunnels []*models.TunnelEntry
}

func NewTunnelEngine(ctx context.Context, cfg []*config.Tunnel) (*TunnelEngine, error) {
	engine := &TunnelEngine{}
	return engine, nil
}
