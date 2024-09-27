/*
 * Copyright (C) 2024 by Jason Figge
 */

package engine

import (
	"context"

	"us.figge.auto-ssh/internal/core/config"
)

type HostEngine struct {
	hosts []*config.Host
}

func NewHostEngine(ctx context.Context, cfg []*config.Host) (*HostEngine, error) {
	engine := &HostEngine{
		hosts: cfg,
	}
	return engine, nil
}

func (h HostEngine) Hosts() []*config.Host {
	return h.hosts
}
