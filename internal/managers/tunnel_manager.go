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
	managerModels "us.figge.auto-ssh/internal/rest/models"
)

var (
	ErrTunnelNotFound = fmt.Errorf("tunnel not found")
	ErrInvalidTunnel  = fmt.Errorf("tunnel definition invalid")
	ErrTunnelRunning  = fmt.Errorf("tunnel already running")
)

type TunnelManager struct {
	tunnels   engineModels.TunnelEngine
	listCache *cache.Cache[string, []*managerModels.TunnelHeader]
}

func NewTunnelManager(ctx context.Context, tunnels engineModels.TunnelEngine) (*TunnelManager, error) {
	manager := &TunnelManager{
		tunnels:   tunnels,
		listCache: cache.NewCache[string, []*managerModels.TunnelHeader](ctx, cache.OptionDefaultTTL(5*time.Minute)),
	}
	return manager, nil
}

func (m *TunnelManager) ListTunnels(
	ctx context.Context,
	input *managerModels.ListTunnelInput,
	opts ...managerModels.TunnelOptionFunc,
) (*managerModels.ListTunnelOutput, error) {
	options := ExtractTunnelOptions(opts)
	output := &managerModels.ListTunnelOutput{}
	var items []*managerModels.TunnelHeader
	if input.More == nil {
		for _, tunnel := range m.tunnels.Tunnels() {
			if tunnelFilter(input.FiltersInput, tunnel) {
				item := &managerModels.TunnelHeader{
					Id:   tunnel.Id(),
					Name: tunnel.Name(),
				}
				if options.Status() {
					item.Status = &config.Status{
						Valid:   tunnel.Valid(),
						Running: tunnel.Running(),
					}
				}
				items = append(items, item)
			}
		}
	} else {
		items, _ = m.listCache.Remove(*input.More)
	}
	output.Items, output.More = Page[*managerModels.TunnelHeader](items, input.PaginationInput, m.listCache)
	output.Count = len(output.Items)
	return output, nil
}

func (m *TunnelManager) GetTunnel(
	ctx context.Context,
	input *managerModels.GetTunnelInput,
	opts ...managerModels.TunnelOptionFunc,
) (*managerModels.GetTunnelOutput, error) {
	options := ExtractTunnelOptions(opts)
	tunnel, ok := m.tunnels.Tunnel(input.Id)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrTunnelNotFound, input.Id)
	}
	output := managerModels.GetTunnelOutput{
		Tunnel: config.Tunnel{
			Id:     tunnel.Id(),
			Name:   tunnel.Name(),
			Local:  tunnel.Local(),
			Remote: tunnel.Remote(),
			Host:   tunnel.Host(),
		},
	}
	if options.Metadata() {
		output.Metadata = tunnel.Metadata()

	}
	if options.Status() {
		output.Status = &config.Status{
			Valid:   tunnel.Valid(),
			Running: tunnel.Running(),
		}

	}
	return &output, nil
}

func (m *TunnelManager) AddTunnel(
	ctx context.Context,
	input *managerModels.AddTunnelInput,
	options ...managerModels.TunnelOptionFunc,
) (*managerModels.AddTunnelOutput, error) {
	return nil, nil
}

func (m *TunnelManager) UpdateTunnel(
	ctx context.Context,
	input *managerModels.UpdateTunnelInput,
	options ...managerModels.TunnelOptionFunc,
) (*managerModels.UpdateTunnelOutput, error) {
	return nil, nil
}

func (m *TunnelManager) RemoveTunnel(
	ctx context.Context,
	input *managerModels.RemoveTunnelInput,
	options ...managerModels.TunnelOptionFunc,
) (*managerModels.RemoveTunnelOutput, error) {
	return nil, nil
}

func (m *TunnelManager) StartTunnel(
	ctx context.Context,
	input *managerModels.StartTunnelInput,
	opts ...managerModels.TunnelOptionFunc,
) (*managerModels.StartTunnelOutput, error) {
	tunnel, ok := m.tunnels.Tunnel(input.Id)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrTunnelNotFound, input.Id)
	}
	if !tunnel.Valid() {
		return nil, fmt.Errorf("%w: %s(%s)", ErrInvalidTunnel, tunnel.Name(), input.Id)
	}
	if strings.EqualFold(tunnel.Running(), "Running") {
		return nil, fmt.Errorf("%w: %s(%s)", ErrTunnelRunning, tunnel.Name(), input.Id)
	}
	tunnel.Start()
	// TODO Move to function and start with Stop
	for range 5 {
		tunnel, _ = m.tunnels.Tunnel(input.Id)
		if tunnel.Running() != "Stopped" {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	output := &managerModels.StartTunnelOutput{Id: input.Id}
	output.Status = &config.Status{
		Valid:   tunnel.Valid(),
		Running: tunnel.Running(),
	}
	return output, nil
}

func (m *TunnelManager) StopTunnel(
	ctx context.Context,
	input *managerModels.StopTunnelInput,
	opts ...managerModels.TunnelOptionFunc,
) (*managerModels.StopTunnelOutput, error) {
	//options := ExtractTunnelOptions(opts)
	tunnel, ok := m.tunnels.Tunnel(input.Id)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrTunnelNotFound, input.Id)
	}
	tunnel.Stop()
	tunnel, _ = m.tunnels.Tunnel(input.Id)
	output := &managerModels.StopTunnelOutput{Id: input.Id}
	output.Status = &config.Status{
		Valid:   tunnel.Valid(),
		Running: tunnel.Running(),
	}
	return output, nil
}

func tunnelFilter(input managerModels.FiltersInput, tunnel engineModels.Tunnel) bool {
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
		case "running":
			match = slices.Contains(filter.Values, tunnel.Running())
		case "metadata.color":
			match = tunnel.Metadata() != nil && slices.Contains(filter.Values, tunnel.Metadata().Color)
		default:
			match = true
		}
		if !match {
			return false
		}
	}
	return true
}
