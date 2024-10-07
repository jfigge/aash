/*
 * Copyright (C) 2024 by Jason Figge
 */

package engine

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"us.figge.auto-ssh/internal/core/config"
	engineModels "us.figge.auto-ssh/internal/resources/models"
)

type TunnelEngine struct {
	tunnelEntries map[string]*TunnelEntry
}

type tunnelData struct {
	*config.Tunnel
	host    engineModels.Host
	valid   bool
	running bool
}

type TunnelEntry struct {
	tunnelData
}

func NewTunnelEngine(ctx context.Context, he engineModels.HostEngine, tunnels []*config.Tunnel) (*TunnelEngine, bool) {
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
	return engine, true
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

func (te *TunnelEngine) StartTunnels(ctx context.Context) (sync.WaitGroup, bool) {
	wg := sync.WaitGroup{}
	listeningChan := make(chan bool)
	for _, tunnel := range te.tunnelEntries {
		wg.Add(1)
		go func(t *TunnelEntry) {
			defer wg.Done()
			t.Open(ctx, listeningChan)
		}(tunnel)
		if !<-listeningChan {
			return wg, false
		}
	}
	return wg, true
}

func (t *TunnelEntry) Open(ctx context.Context, listeningChan chan bool) {

}

func (t *TunnelEntry) Validate(he engineModels.HostEngine) bool {
	t.tunnelData.Name = strings.TrimSpace(t.tunnelData.Name)
	if t.tunnelData.Name == "" {
		fmt.Printf("  Error - tunnel name cannot be blank\n")
		t.valid = false
	}
	if t.tunnelData.Remote == nil || t.tunnelData.Remote.IsBlank() {
		fmt.Printf("  Error - tunnel (%s) requires a forward address\n", t.tunnelData.Name)
		t.valid = false
	} else if !t.tunnelData.Remote.Validate("tunnel", t.tunnelData.Name, "forward address", true, false) {
		t.valid = false
	}

	if (t.tunnelData.Local == nil || t.tunnelData.Local.IsBlank()) && t.tunnelData.Remote != nil && t.tunnelData.Remote.IsValid() {
		fmt.Printf("  Warn  - tunnel (%s) Local entrance undefined. Defaulting to 127.0.0.1:%d\n", t.tunnelData.Name, t.tunnelData.Remote.Port())
		t.tunnelData.Local = config.NewAddress(fmt.Sprintf("127.0.0.1:%d", t.tunnelData.Remote.Port()))
	}
	if t.tunnelData.Local == nil || t.tunnelData.Local.IsBlank() {
		fmt.Printf("  Error - tunnel (%s) missing a local address that cannot be derived\n", t.tunnelData.Name)
	} else if !t.tunnelData.Local.Validate("tunnel", t.tunnelData.Name, "local address", true, false) {
		t.valid = false
	}

	t.tunnelData.Host = strings.TrimSpace(t.tunnelData.Host)
	if t.tunnelData.Host == "" {
		fmt.Printf("  Info  - tunnel (%s) exits on the local host\n", t.tunnelData.Name)
	} else if host, ok := he.Host(t.tunnelData.Host); !ok {
		fmt.Printf("  Error - tunnel (%s) remote host (%s) undefined\n", t.tunnelData.Name, t.tunnelData.Host)
		t.valid = false
	} else if t.valid {
		he.MarkInUse(host.Name())
		t.host = host
	}

	if config.VerboseFlag && t.valid {
		fmt.Printf("  Info  - tunnel (%s) validated\n", t.tunnelData.Name)
	}

	//t.stats = &TunnelStats{
	//	Name: t.Name,
	//	Port: t.Local.Port(),
	//}

	return t.valid
}

func (t *TunnelEntry) Id() string {
	return t.tunnelData.Id
}

func (t *TunnelEntry) Name() string {
	return t.tunnelData.Name
}

func (t *TunnelEntry) Local() *config.Address {
	return t.tunnelData.Local
}

func (t *TunnelEntry) Remote() *config.Address {
	return t.tunnelData.Remote
}

func (t *TunnelEntry) Host() string {
	return t.tunnelData.Host
}

func (t *TunnelEntry) Valid() bool {
	return t.tunnelData.valid
}

func (t *TunnelEntry) Metadata() *config.Metadata {
	return t.tunnelData.Metadata
}
