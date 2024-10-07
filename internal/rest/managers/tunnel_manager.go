/*
 * Copyright (C) 2024 by Jason Figge
 */

package managers

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"us.figge.auto-ssh/internal/core/config"
	"us.figge.auto-ssh/internal/core/utils/cache"
	engineModels "us.figge.auto-ssh/internal/resources/models"
	"us.figge.auto-ssh/internal/rest/models"
)

var (
	ErrTunnelNotFound = fmt.Errorf("tunnel not found")
)

type TunnelManager struct {
	tunnels   engineModels.TunnelEngine
	listCache *cache.Cache[string, []*models.TunnelHeader]
}

func NewTunnelManager(ctx context.Context, tunnels engineModels.TunnelEngine) (*TunnelManager, error) {
	manager := &TunnelManager{
		tunnels:   tunnels,
		listCache: cache.NewCache[string, []*models.TunnelHeader](ctx, cache.OptionDefaultTTL(5*time.Minute)),
	}
	return manager, nil
}

func (m *TunnelManager) ListTunnels(
	ctx context.Context,
	input *models.ListTunnelInput,
	options ...models.TunnelOptionFunc,
) (*models.ListTunnelOutput, error) {
	output := &models.ListTunnelOutput{}
	var items []*models.TunnelHeader
	if input.More == nil {
		for _, tunnel := range m.tunnels.Tunnels() {
			if tunnelFilter(input.FiltersInput, tunnel) {
				items = append(items, &models.TunnelHeader{Id: tunnel.Id(), Name: tunnel.Name(), Valid: tunnel.Valid()})
			}
		}
	} else {
		items, _ = m.listCache.Remove(*input.More)
	}
	output.Items, output.More = Page[*models.TunnelHeader](items, input.PaginationInput, m.listCache)
	output.Count = len(output.Items)
	return output, nil
}

func (m *TunnelManager) GetTunnel(
	ctx context.Context,
	input *models.GetTunnelInput,
	options ...models.TunnelOptionFunc,
) (*models.GetTunnelOutput, error) {
	tunnel, ok := m.tunnels.Tunnel(input.Id)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrTunnelNotFound, input.Id)
	}
	output := models.GetTunnelOutput{
		Tunnel: config.Tunnel{
			Id:       tunnel.Id(),
			Name:     tunnel.Name(),
			Local:    tunnel.Local(),
			Remote:   tunnel.Remote(),
			Host:     tunnel.Host(),
			Metadata: tunnel.Metadata(),
		},
	}
	return &output, nil
}

func (m *TunnelManager) AddTunnel(
	ctx context.Context,
	input *models.AddTunnelInput,
	options ...models.TunnelOptionFunc,
) (*models.AddTunnelOutput, error) {
	return nil, nil
}

func (m *TunnelManager) UpdateTunnel(
	ctx context.Context,
	input *models.UpdateTunnelInput,
	options ...models.TunnelOptionFunc,
) (*models.UpdateTunnelOutput, error) {
	return nil, nil
}

func (m *TunnelManager) RemoveTunnel(
	ctx context.Context,
	input *models.RemoveTunnelInput,
	options ...models.TunnelOptionFunc,
) (*models.RemoveTunnelOutput, error) {
	return nil, nil
}

func (m *TunnelManager) StartTunnel(
	ctx context.Context,
	input *models.StartTunnelInput,
	options ...models.TunnelOptionFunc,
) (*models.StartTunnelOutput, error) {
	return nil, nil
}

func (m *TunnelManager) StopTunnel(
	ctx context.Context,
	input *models.StopTunnelInput,
	options ...models.TunnelOptionFunc,
) (*models.StopTunnelOutput, error) {
	return nil, nil
}

func tunnelFilter(input models.FiltersInput, tunnel engineModels.Tunnel) bool {
	for _, filter := range input.Filters {
		match := false
		switch strings.ToLower(filter.Key) {
		case "id":
			match = slices.Contains(filter.Values, tunnel.Id())
		case "name":
			match = slices.Contains(filter.Values, tunnel.Name())
		case "tags":
			match = contains(filter.Values, tunnel.Metadata().Tags)
		case "local":
			match = slices.Contains(filter.Values, tunnel.Local().String())
		case "remote":
			match = slices.Contains(filter.Values, tunnel.Remote().String())
		case "host":
			match = slices.Contains(filter.Values, tunnel.Host())
		case "valid":
			match = slices.Contains(filter.Values, strconv.FormatBool(tunnel.Valid()))
		case "metadata.color":
			match = tunnel.Metadata() != nil && slices.Contains(filter.Values, tunnel.Metadata().Color)
		}
		if !match {
			return false
		}
	}
	return true
}
