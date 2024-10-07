/*
 * Copyright (C) 2024 by Jason Figge
 */

package engine

import (
	"context"
	"slices"

	"us.figge.auto-ssh/internal/core/config"
	engineModels "us.figge.auto-ssh/internal/resources/models"
)

type HostEngine struct {
	hostEntries map[string]*HostEntry
}

type hostData struct {
	*config.Host
	valid bool
	inUse bool
}
type HostEntry struct {
	hostData
}

func NewHostEngine(ctx context.Context, hosts []*config.Host) (*HostEngine, bool) {
	engine := &HostEngine{
		hostEntries: make(map[string]*HostEntry),
	}
	success := true
	for _, cfgHost := range hosts {
		host := &HostEntry{
			hostData: hostData{
				Host:  cfgHost,
				valid: false,
				inUse: false,
			},
		}
		engine.hostEntries[cfgHost.Id] = host
	}
	return engine, success
}

func (he *HostEngine) Hosts() []engineModels.Host {
	hosts := make([]engineModels.Host, 0, len(he.hostEntries))
	for _, hostEntry := range he.hostEntries {
		hosts = append(hosts, hostEntry)
	}
	return hosts
}

func (he *HostEngine) Host(id string) (engineModels.Host, bool) {
	host, ok := he.hostEntries[id]
	return host, ok
}

func (he *HostEngine) KnownHosts() []string {
	knownHosts := make([]string, 0)
	for _, hostEntry := range he.hostEntries {
		if hostEntry.hostData.KnownHosts != "" && !slices.Contains(knownHosts, hostEntry.hostData.KnownHosts) {
			knownHosts = append(knownHosts, hostEntry.hostData.KnownHosts)
		}
	}
	return knownHosts
}

func (he *HostEngine) MarkInUse(name string) {
	if hostEntry, ok := he.hostEntries[name]; ok {
		hostEntry.inUse = true
	}
}

func (h *HostEntry) Id() string {
	return h.hostData.Id
}
func (h *HostEntry) Name() string {
	return h.hostData.Name
}
func (h *HostEntry) Remote() *config.Address {
	return h.hostData.Remote
}
func (h *HostEntry) Username() string {
	return h.hostData.Username
}
func (h *HostEntry) Identity() string {
	return h.hostData.Identity
}
func (h *HostEntry) KnownHosts() string {
	return h.hostData.KnownHosts
}
func (h *HostEntry) JumpHost() string {
	return h.hostData.JumpHost
}
func (h *HostEntry) Valid() bool {
	return h.hostData.valid
}
func (h *HostEntry) Metadata() *config.Metadata {
	return h.hostData.Metadata
}
