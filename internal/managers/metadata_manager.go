/*
 * Copyright (C) 2024 by Jason Figge
 */

package managers

import (
	"context"
	"slices"

	engineModels "us.figge.auto-ssh/internal/resources/models"
	managerModels "us.figge.auto-ssh/internal/rest/models"
)

type MetadataManager struct {
	tunnels engineModels.TunnelEngine
}

func NewMetadataManager(ctx context.Context, tunnels engineModels.TunnelEngine) (*MetadataManager, error) {
	manager := &MetadataManager{
		tunnels: tunnels,
	}
	return manager, nil
}

func (m MetadataManager) ListStates(ctx context.Context, options ...managerModels.MetadataOptionFunc) (*managerModels.ListMetadataStatesOutput, error) {
	output := &managerModels.ListMetadataStatesOutput{}
	for _, enum := range engineModels.RunningEnums() {
		output.States = append(output.States, enum)
	}
	return output, nil
}

func (m MetadataManager) ListTags(ctx context.Context, input *managerModels.ListMetadataTagsInput, options ...managerModels.MetadataOptionFunc) (*managerModels.ListMetadataTagsOutput, error) {
	output := &managerModels.ListMetadataTagsOutput{}
	for _, tunnel := range m.tunnels.Tunnels() {
		if tunnelFilter(input.FiltersInput, tunnel) {
			for _, tag := range tunnel.Metadata().Tags {
				if !slices.Contains(output.Tags, tag) {
					output.Tags = append(output.Tags, tag)
				}
			}
		}
	}
	return output, nil
}
