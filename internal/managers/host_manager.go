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
	ErrHostNotFound = fmt.Errorf("host not found")
)

type HostManager struct {
	hosts               engineModels.HostEngine
	listHostHeaderCache *cache.Cache[string, []*managerModels.HostHeader]
	listKnownHostCache  *cache.Cache[string, []*managerModels.KnownHost]
}

func NewHostManager(ctx context.Context, hosts engineModels.HostEngine) (*HostManager, error) {
	manager := &HostManager{
		hosts:               hosts,
		listHostHeaderCache: cache.NewCache[string, []*managerModels.HostHeader](ctx, cache.OptionDefaultTTL(5*time.Minute)),
		listKnownHostCache:  cache.NewCache[string, []*managerModels.KnownHost](ctx, cache.OptionDefaultTTL(5*time.Minute)),
	}
	return manager, nil
}

func (m *HostManager) ListHosts(
	ctx context.Context,
	input *managerModels.ListHostInput,
	options ...managerModels.HostOptionFunc,
) (*managerModels.ListHostOutput, error) {
	output := &managerModels.ListHostOutput{}
	var items []*managerModels.HostHeader
	if input.More == nil {
		for _, host := range m.hosts.Hosts() {
			if hostFilter(input.FiltersInput, host) {
				items = append(items, &managerModels.HostHeader{Id: host.Id(), Name: host.Name(), Valid: host.Valid()})
			}
		}
	} else {
		items, _ = m.listHostHeaderCache.Remove(*input.More)
	}
	output.Items, output.More = Page[*managerModels.HostHeader](items, input.PaginationInput, m.listHostHeaderCache)
	output.Count = len(output.Items)
	return output, nil
}

func (m *HostManager) GetHost(
	ctx context.Context,
	input *managerModels.GetHostInput,
	options ...managerModels.HostOptionFunc,
) (*managerModels.GetHostOutput, error) {
	host, ok := m.hosts.Host(input.Id)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrHostNotFound, input.Id)
	}
	output := managerModels.GetHostOutput{
		Host: config.Host{
			Id:         host.Id(),
			Name:       host.Name(),
			Remote:     host.Remote(),
			Username:   host.Username(),
			Identity:   host.Identity(),
			KnownHosts: host.KnownHosts(),
			JumpHost:   host.JumpHost(),
			Metadata:   host.Metadata(),
		},
	}
	return &output, nil
}

func (m *HostManager) AddHost(
	ctx context.Context,
	input *managerModels.AddHostInput,
	options ...managerModels.HostOptionFunc,
) (*managerModels.AddHostOutput, error) {
	return nil, nil
}

func (m *HostManager) UpdateHost(
	ctx context.Context,
	input *managerModels.UpdateHostInput,
	options ...managerModels.HostOptionFunc,
) (*managerModels.UpdateHostOutput, error) {
	return nil, nil
}

func (m *HostManager) RemoveHost(
	ctx context.Context,
	input *managerModels.RemoveHostInput,
	options ...managerModels.HostOptionFunc,
) (*managerModels.RemoveHostOutput, error) {
	return nil, nil
}

func (m *HostManager) ListKnownHosts(
	ctx context.Context,
	input *managerModels.ListKnownHostsInput,
	options ...managerModels.HostOptionFunc,
) (*managerModels.ListKnownHostsOutput, error) {
	output := &managerModels.ListKnownHostsOutput{}
	var items []*managerModels.KnownHost
	if input.More == nil {
		for _, knownHost := range m.hosts.KnownHosts() {
			items = append(items, &managerModels.KnownHost{File: knownHost})
		}
	} else {
		items, _ = m.listKnownHostCache.Remove(*input.More)
	}
	output.Items, output.More = Page[*managerModels.KnownHost](items, input.PaginationInput, m.listKnownHostCache)
	output.Count = len(output.Items)
	return output, nil
}

func hostFilter(input managerModels.FiltersInput, host engineModels.Host) bool {
	for _, filter := range input.Filters {
		match := false
		switch strings.ToLower(filter.Key) {
		case "id":
			match = slices.Contains(filter.Values, host.Id())
		case "name":
			match = slices.Contains(filter.Values, host.Name())
		case "tags":
			match = contains(filter.Values, host.Metadata().Tags)
		case "address":
			match = slices.Contains(filter.Values, host.Remote().String())
		case "username":
			match = slices.Contains(filter.Values, host.Username())
		case "identity":
			match = slices.Contains(filter.Values, host.Identity())
		case "knownHosts":
			match = slices.Contains(filter.Values, host.KnownHosts())
		case "host", "jumpHost":
			match = slices.Contains(filter.Values, host.JumpHost())
		case "valid":
			match = slices.Contains(filter.Values, strconv.FormatBool(host.Valid()))
		case "metadata.color":
			match = host.Metadata() != nil && slices.Contains(filter.Values, host.Metadata().Color)
		default:
			match = true
		}
		if !match {
			return false
		}
	}
	return true
}
