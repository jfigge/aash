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

type Engine struct {
	lock          *sync.Mutex
	tunnelEntries map[string]*Entry
}

func NewEngine(ctx context.Context, he engineModels.HostEngineInternal, tunnels []*config.Tunnel) *Engine {
	engine := &Engine{
		lock:          &sync.Mutex{},
		tunnelEntries: make(map[string]*Entry),
	}
	for _, cfgTunnel := range tunnels {
		if _, ok := engine.tunnelEntries[cfgTunnel.Name]; ok {
			fmt.Printf("  Error - tunnel name (%s) redfined\n", cfgTunnel.Name)
			continue
		}
		tunnel := &Entry{
			tunnelData: &tunnelData{
				Tunnel: cfgTunnel,
			},
		}
		tunnel.Status = &config.Status{
			Running: "Stopped",
			Valid:   true,
		}
		tunnel.Validate(he)
		engine.tunnelEntries[tunnel.tunnelData.Id] = tunnel
	}
	he.ValidateJumpHosts(engine.tunnelEntries)
	return engine
}

func (te *Engine) Tunnels() []engineModels.Tunnel {
	tunnels := make([]engineModels.Tunnel, 0, len(te.tunnelEntries))
	for _, tunnelEntry := range te.tunnelEntries {
		tunnels = append(tunnels, tunnelEntry)
	}
	return tunnels
}

func (te *Engine) Tunnel(id string) (engineModels.Tunnel, bool) {
	tunnel, ok := te.tunnelEntries[id]
	return tunnel, ok
}

func (te *Engine) StartTunnels(ctx context.Context, statsEngine engineModels.StatsEngine, wg *sync.WaitGroup) {
	go te.waitForTermination(ctx)
	for _, tunnel := range te.tunnelEntries {
		statsEntry := statsEngine.NewEntry()
		tunnel.init(ctx, statsEntry, wg)
		if !tunnel.Valid() {
			continue
		}
		tunnel.Start()
	}
}

func (te *Engine) waitForTermination(ctx context.Context) {
	<-ctx.Done()
	te.lock.Lock()
	defer te.lock.Unlock()
	for _, tunnel := range te.tunnelEntries {
		tunnel.Stop()
	}
}
