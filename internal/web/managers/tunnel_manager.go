/*
 * Copyright (C) 2024 by Jason Figge
 */

package managers

import (
	"context"

	engineModels "us.figge.auto-ssh/internal/resources/models"
	"us.figge.auto-ssh/internal/web/models"
)

type TunnelManager struct {
	tunnels engineModels.Tunnel
}

func NewTunnelManager(ctx context.Context, tunnels engineModels.Tunnel) (*TunnelManager, error) {
	manager := &TunnelManager{
		tunnels: tunnels,
	}
	return manager, nil
}

func (m *TunnelManager) List(
	ctx context.Context,
	input *models.ListTunnelInput,
	options ...models.TunnelOptionFunc,
) (*models.ListTunnelOutput, error) {
	return nil, nil
}

func (m *TunnelManager) Get(
	ctx context.Context,
	input *models.GetTunnelInput,
	options ...models.TunnelOptionFunc,
) (*models.GetTunnelOutput, error) {
	return nil, nil
}

func (m *TunnelManager) Add(
	ctx context.Context,
	input *models.AddTunnelInput,
	options ...models.TunnelOptionFunc,
) (*models.AddTunnelOutput, error) {
	return nil, nil
}

func (m *TunnelManager) Update(
	ctx context.Context,
	input *models.UpdateTunnelInput,
	options ...models.TunnelOptionFunc,
) (*models.UpdateTunnelOutput, error) {
	return nil, nil
}

func (m *TunnelManager) Remove(
	ctx context.Context,
	input *models.RemoveTunnelInput,
	options ...models.TunnelOptionFunc,
) (*models.RemoveTunnelOutput, error) {
	return nil, nil
}

func (m *TunnelManager) Start(
	ctx context.Context,
	input *models.StartTunnelInput,
	options ...models.TunnelOptionFunc,
) (*models.StartTunnelOutput, error) {
	return nil, nil
}

func (m *TunnelManager) Stop(
	ctx context.Context,
	input *models.StopTunnelInput,
	options ...models.TunnelOptionFunc,
) (*models.StopTunnelOutput, error) {
	return nil, nil
}
