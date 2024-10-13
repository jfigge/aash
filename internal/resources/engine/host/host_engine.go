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

type Engine struct {
	hostEntries map[string]*Entry
	identityMap map[string]ssh.Signer
	hostKeysMap map[string]*HostKeyManager
}

func NewEngine(ctx context.Context, hosts []*config.Host) *Engine {
	engine := &Engine{
		hostEntries: make(map[string]*Entry),
		identityMap: make(map[string]ssh.Signer),
		hostKeysMap: make(map[string]*HostKeyManager),
	}
	for _, cfgHost := range hosts {
		if _, ok := engine.hostEntries[cfgHost.Name]; ok {
			fmt.Printf("  Error - host name (%s) redfined\n", cfgHost.Name)
			continue
		}
		host := &Entry{
			hostData: &hostData{
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

func (he *Engine) Hosts() []engineModels.Host {
	hosts := make([]engineModels.Host, 0, len(he.hostEntries))
	for _, hostEntry := range he.hostEntries {
		hosts = append(hosts, hostEntry)
	}
	return hosts
}

func (he *Engine) Host(id string) (engineModels.Host, bool) {
	host, ok := he.hostEntries[id]
	return host, ok
}

func (he *Engine) KnownHosts() []string {
	knownHosts := make([]string, 0)
	for _, hostEntry := range he.hostEntries {
		if hostEntry.hostData.KnownHosts != "" && !slices.Contains(knownHosts, hostEntry.hostData.KnownHosts) {
			knownHosts = append(knownHosts, hostEntry.hostData.KnownHosts)
		}
	}
	return knownHosts
}
