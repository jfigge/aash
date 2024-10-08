/*
 * Copyright (C) 2024 by Jason Figge
 */

package host

import (
	"context"
	"fmt"
	"slices"

	"golang.org/x/crypto/ssh"
	"us.figge.auto-ssh/internal/core/config"
	engineModels "us.figge.auto-ssh/internal/resources/models"
)

type HostEngine struct {
	hostEntries map[string]*HostEntry
	identityMap map[string]ssh.Signer
	hostKeysMap map[string]*HostKeyManager
}

func NewHostEngine(ctx context.Context, hosts []*config.Host) *HostEngine {
	engine := &HostEngine{
		hostEntries: make(map[string]*HostEntry),
		identityMap: make(map[string]ssh.Signer),
		hostKeysMap: make(map[string]*HostKeyManager),
	}
	for _, cfgHost := range hosts {
		if _, ok := engine.hostEntries[cfgHost.Name]; ok {
			fmt.Printf("  Error - host name (%s) redfined\n", cfgHost.Name)
			continue
		}
		host := &HostEntry{
			hostData: hostData{
				Host:  cfgHost,
				valid: true,
				inUse: false,
			},
		}
		host.Validate("", engine.identityMap, engine.hostKeysMap)
		engine.hostEntries[cfgHost.Id] = host
	}
	return engine
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
