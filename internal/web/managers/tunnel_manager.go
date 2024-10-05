/*
 * Copyright (C) 2024 by Jason Figge
 */

package managers

import (
	"context"
	"slices"
	"strings"
	"time"

	"us.figge.auto-ssh/internal/cache"
	"us.figge.auto-ssh/internal/core/config"
	engineModels "us.figge.auto-ssh/internal/resources/models"
	"us.figge.auto-ssh/internal/web/models"
)

type TunnelManager struct {
	tunnels   engineModels.Tunnel
	listCache *cache.Cache[string, []*config.Tunnel]
}

func NewTunnelManager(ctx context.Context, tunnels engineModels.Tunnel) (*TunnelManager, error) {
	manager := &TunnelManager{
		tunnels: tunnels,
		listCache: cache.NewCache[string, []*config.Tunnel](
			ctx,
			cache.OptionDefaultTTL[string, []*config.Tunnel](5*time.Minute),
		),
	}
	return manager, nil
}

func (m *TunnelManager) List(
	ctx context.Context,
	input *models.ListTunnelInput,
	options ...models.TunnelOptionFunc,
) (*models.ListTunnelOutput, error) {
	output := &models.ListTunnelOutput{}
	var items []*config.Tunnel
	if input.More == nil {
		for _, tunnel := range m.tunnels.Tunnels() {
			if tunnelFilter(input.Filter, tunnel) {
				items = append(items, tunnel)
			}
		}
	} else {
		items, _ = m.listCache.Remove(*input.More)
	}
	output.Items, output.More = models.Page[config.Tunnel](items, input.Pagination)
	output.Count = len(output.Items)
	return output, nil
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

func tunnelFilter(filters []*models.Filter, tunnel *config.Tunnel) bool {
	for _, filter := range filters {
		match := false
		switch strings.ToLower(filter.Key) {
		case "name":
			match = slices.Contains(filter.Values, tunnel.Name)
		case "group":
			match = slices.Contains(filter.Values, tunnel.Group)
		case "local":
			match = slices.Contains(filter.Values, tunnel.Local)
		case "remote":
			match = slices.Contains(filter.Values, tunnel.Remote)
		case "jump-host", "jumpHost":
			match = slices.Contains(filter.Values, tunnel.JumpHost)
		case "metadata.color":
			match = tunnel.Metadata != nil && slices.Contains(filter.Values, tunnel.Metadata.Color)
		}
		if !match {
			return false
		}
	}
	return true
}
