/*
 * Copyright (C) 2024 by Jason Figge
 */

package tunnel

import (
	"context"
	"fmt"
	"sync"

	"us.figge.auto-ssh/internal/core/config"
	engineModels "us.figge.auto-ssh/internal/resources/models"
)

type TunnelEngine struct {
	tunnelEntries map[string]*TunnelEntry
}

func NewTunnelEngine(ctx context.Context, he engineModels.HostEngineInternal, tunnels []*config.Tunnel) *TunnelEngine {
	engine := &TunnelEngine{
		tunnelEntries: make(map[string]*TunnelEntry),
	}
	for _, cfgTunnel := range tunnels {
		if _, ok := engine.tunnelEntries[cfgTunnel.Name]; ok {
			fmt.Printf("  Error - tunnel name (%s) redfined\n", cfgTunnel.Name)
			continue
		}
		tunnel := &TunnelEntry{
			tunnelData: tunnelData{
				Tunnel:  cfgTunnel,
				valid:   true,
				running: false,
			},
		}
		tunnel.Validate(he)
		engine.tunnelEntries[tunnel.tunnelData.Id] = tunnel
	}
	return engine
}

func (te *TunnelEngine) Tunnels() []engineModels.Tunnel {
	tunnels := make([]engineModels.Tunnel, 0, len(te.tunnelEntries))
	for _, tunnelEntry := range te.tunnelEntries {
		tunnels = append(tunnels, tunnelEntry)
	}
	return tunnels
}

func (te *TunnelEngine) Tunnel(id string) (engineModels.Tunnel, bool) {
	tunnel, ok := te.tunnelEntries[id]
	return tunnel, ok
}

func (te *TunnelEngine) StartTunnels(ctx context.Context, wg *sync.WaitGroup) {
	for _, tunnel := range te.tunnelEntries {
		if !tunnel.Valid() {
			continue
		}
		wg.Add(1)
		go func(t *TunnelEntry) {
			defer wg.Done()
			t.Open(ctx)
		}(tunnel)
	}
}
