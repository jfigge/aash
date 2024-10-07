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
	ErrHostNotFound = fmt.Errorf("host not found")
)

type HostManager struct {
	hosts               engineModels.HostEngine
	listHostHeaderCache *cache.Cache[string, []*models.HostHeader]
	listKnownHostCache  *cache.Cache[string, []*models.KnownHost]
}

func NewHostManager(ctx context.Context, hosts engineModels.HostEngine) (*HostManager, error) {
	manager := &HostManager{
		hosts:               hosts,
		listHostHeaderCache: cache.NewCache[string, []*models.HostHeader](ctx, cache.OptionDefaultTTL(5*time.Minute)),
		listKnownHostCache:  cache.NewCache[string, []*models.KnownHost](ctx, cache.OptionDefaultTTL(5*time.Minute)),
	}
	return manager, nil
}

func (m *HostManager) ListHosts(
	ctx context.Context,
	input *models.ListHostInput,
	options ...models.HostOptionFunc,
) (*models.ListHostOutput, error) {
	output := &models.ListHostOutput{}
	var items []*models.HostHeader
	if input.More == nil {
		for _, host := range m.hosts.Hosts() {
			if hostFilter(input.FiltersInput, host) {
				items = append(items, &models.HostHeader{Id: host.Id(), Name: host.Name(), Valid: host.Valid()})
			}
		}
	} else {
		items, _ = m.listHostHeaderCache.Remove(*input.More)
	}
	output.Items, output.More = Page[*models.HostHeader](items, input.PaginationInput, m.listHostHeaderCache)
	output.Count = len(output.Items)
	return output, nil
}

func (m *HostManager) GetHost(
	ctx context.Context,
	input *models.GetHostInput,
	options ...models.HostOptionFunc,
) (*models.GetHostOutput, error) {
	host, ok := m.hosts.Host(input.Id)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrHostNotFound, input.Id)
	}
	output := models.GetHostOutput{
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
	input *models.AddHostInput,
	options ...models.HostOptionFunc,
) (*models.AddHostOutput, error) {
	return nil, nil
}

func (m *HostManager) UpdateHost(
	ctx context.Context,
	input *models.UpdateHostInput,
	options ...models.HostOptionFunc,
) (*models.UpdateHostOutput, error) {
	return nil, nil
}

func (m *HostManager) RemoveHost(
	ctx context.Context,
	input *models.RemoveHostInput,
	options ...models.HostOptionFunc,
) (*models.RemoveHostOutput, error) {
	return nil, nil
}

func (m *HostManager) ListKnownHosts(
	ctx context.Context,
	input *models.ListKnownHostsInput,
	options ...models.HostOptionFunc,
) (*models.ListKnownHostsOutput, error) {
	output := &models.ListKnownHostsOutput{}
	var items []*models.KnownHost
	if input.More == nil {
		for _, knownHost := range m.hosts.KnownHosts() {
			items = append(items, &models.KnownHost{File: knownHost})
		}
	} else {
		items, _ = m.listKnownHostCache.Remove(*input.More)
	}
	output.Items, output.More = Page[*models.KnownHost](items, input.PaginationInput, m.listKnownHostCache)
	output.Count = len(output.Items)
	return output, nil
}

func hostFilter(input models.FiltersInput, host engineModels.Host) bool {
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

		}
		if !match {
			return false
		}
	}
	return true
}
