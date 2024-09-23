/*
 * Copyright (C) 2024 by Jason Figge
 */

package managers

import (
	"context"

	"us.figge.auto-ssh/internal/web/models"
)

type HostOptionFunc func(options *HostOptions)
type HostOptions struct{}
type Host struct{}

func NewHostManager() (*Host, error) {
	manager := &Host{}
	return manager, nil
}

func (m *Host) List(
	ctx context.Context,
	input *models.ListHostInput,
	options ...HostOptionFunc,
) (*models.ListHostOutput, error) {
	return nil, nil
}

func (m *Host) Get(
	ctx context.Context,
	input *models.GetHostInput,
	options ...HostOptionFunc,
) (*models.GetHostOutput, error) {
	return nil, nil
}

func (m *Host) Add(
	ctx context.Context,
	input *models.AddHostInput,
	options ...HostOptionFunc,
) (*models.AddHostOutput, error) {
	return nil, nil
}

func (m *Host) Remove(
	ctx context.Context,
	input *models.RemoveHostInput,
	options ...HostOptionFunc,
) (*models.RemoveHostOutput, error) {
	return nil, nil
}
