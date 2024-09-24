/*
 * Copyright (C) 2024 by Jason Figge
 */

package managers

import (
	"context"

	engineModels "us.figge.auto-ssh/internal/resources/models"
	"us.figge.auto-ssh/internal/web/models"
)

type HostManager struct {
	hosts engineModels.Host
}

func NewHostManager(ctx context.Context, hosts engineModels.Host) (*HostManager, error) {
	manager := &HostManager{
		hosts: hosts,
	}
	return manager, nil
}

func (m *HostManager) List(
	ctx context.Context,
	input *models.ListHostInput,
	options ...models.HostOptionFunc,
) (*models.ListHostOutput, error) {
	return nil, nil
}

func (m *HostManager) Get(
	ctx context.Context,
	input *models.GetHostInput,
	options ...models.HostOptionFunc,
) (*models.GetHostOutput, error) {
	return nil, nil
}

func (m *HostManager) Add(
	ctx context.Context,
	input *models.AddHostInput,
	options ...models.HostOptionFunc,
) (*models.AddHostOutput, error) {
	return nil, nil
}

func (m *HostManager) Update(
	ctx context.Context,
	input *models.UpdateHostInput,
	options ...models.HostOptionFunc,
) (*models.UpdateHostOutput, error) {
	return nil, nil
}

func (m *HostManager) Remove(
	ctx context.Context,
	input *models.RemoveHostInput,
	options ...models.HostOptionFunc,
) (*models.RemoveHostOutput, error) {
	return nil, nil
}
