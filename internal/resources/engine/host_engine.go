/*
 * Copyright (C) 2024 by Jason Figge
 */

package engine

import (
	"context"

	"us.figge.auto-ssh/internal/core/config"
	"us.figge.auto-ssh/internal/resources/models"
)

type HostEngine struct {
	hosts []*models.HostEntry
}

func NewHostEngine(ctx context.Context, cfg []*config.Host) (*HostEngine, error) {
	engine := &HostEngine{}
	return engine, nil
}
