/*
 * Copyright (C) 2024 by Jason Figge
 */

package managers

import (
	"context"

	"us.figge.auto-ssh/internal/web/models"
)

type TunnelOptionFunc func(options *TunnelOptions)
type TunnelOptions struct{}
type Tunnel struct{}

func NewTunnelManager() (*Tunnel, error) {
	manager := &Tunnel{}
	return manager, nil
}

func (m *Tunnel) List(
	ctx context.Context,
	input *models.ListTunnelInput,
	options ...TunnelOptionFunc,
) (*models.ListTunnelOutput, error) {
	return nil, nil
}

func (m *Tunnel) Get(
	ctx context.Context,
	input *models.GetTunnelInput,
	options ...TunnelOptionFunc,
) (*models.GetTunnelOutput, error) {
	return nil, nil
}

func (m *Tunnel) Add(
	ctx context.Context,
	input *models.AddTunnelInput,
	options ...TunnelOptionFunc,
) (*models.AddTunnelOutput, error) {
	return nil, nil
}

func (m *Tunnel) Remove(
	ctx context.Context,
	input *models.RemoveTunnelInput,
	options ...TunnelOptionFunc,
) (*models.RemoveTunnelOutput, error) {
	return nil, nil
}

func (m *Tunnel) Start(
	ctx context.Context,
	input *models.StartTunnelInput,
	options ...TunnelOptionFunc,
) (*models.StartTunnelOutput, error) {
	return nil, nil
}

func (m *Tunnel) Stop(
	ctx context.Context,
	input *models.StopTunnelInput,
	options ...TunnelOptionFunc,
) (*models.StopTunnelOutput, error) {
	return nil, nil
}

func (m *Tunnel) Restart(
	ctx context.Context,
	input *models.RestartTunnelInput,
	options ...TunnelOptionFunc,
) (*models.RestartTunnelOutput, error) {
	return nil, nil
}
