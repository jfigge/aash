/*
 * Copyright (C) 2024 by Jason Figge
 */

package engine

import (
	"context"

	"us.figge.auto-ssh/internal/core/config"
)

type Host struct {
	Name       string
	Group      string
	Address    string
	Username   string
	Identity   string
	KnownHosts string
	JumpHost   string
	Metadata   *config.Metadata
}

func NewHostEngine(ctx context.Context, cfg config.Host) (*Host, error) {
	engine := &Host{}
	return engine, nil
}
