/*
 * Copyright (C) 2024 by Jason Figge
 */

package host

import (
	"context"
	"fmt"
	"net"
	"slices"
	"sync"

	"golang.org/x/crypto/ssh"
	"us.figge.auto-ssh/internal/core/config"
	engineTunnel "us.figge.auto-ssh/internal/resources/engine/tunnel"
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

func (he *Engine) ValidateJumpHosts(tunnelEntries map[string]*engineTunnel.Entry) {

	var freePortListeners []net.Listener
	defer func() {
		for _, listener := range freePortListeners {
			_ = listener.Close()
		}
	}()
	for _, h := range he.hostEntries {
		if h.JumpHost() != "" && h.inUse {
			jumpHost, ok := he.hostEntries[h.JumpHost()]
			if !ok {
				fmt.Printf("  Error - host (%s) jump_host (%s) is not defined\n", h.Name(), h.JumpHost())
				h.valid = false
				continue
			}
			if jumpHost.JumpHost() != "" {
				fmt.Printf("  Error - host (%s) requires multi-host jumps and is not supported", h.Name)
				h.valid = false
				continue
			}
				listener, port, found := freePort()
				if !found {
					h.valid = false
					continue
				} else {
					jumpTunnel := engineTunnel.NewTunnel(
						fmt.Sprintf("%s jumphost", jumpHost.Name()),
						h.JumpHost(),
						config.NewAddress(fmt.Sprintf("127.0.0.1:%d", port)),
						h.Remote(),
						)
						//Host:    h.JumpHost,
						//Forward: h.Address,
					}
					//if jumpTunnel.Validate(he, sm.UpdateChannel()) {
					//	jumpTunnel.stats.JumpTunnel = true
					//	h.Address = jumpTunnel.Local
					//	freePortListeners = append(freePortListeners, listener)
					//	tunnelMap[jumpTunnel.Name] = jumpTunnel
					//	sm.AddTunnelStats(jumpTunnel.stats)
					//} else {
					//	valid = false
					//}

			}
		}
	}

	if valid {
		var unused []string
		for name, host := range he./hostMap {
			if !host.isHost && !host.isJumpHost {
				log.Printf("  Info  - host (%s) is unused\n", name)
				unused = append(unused, name)
			}
		}
		for _, name := range unused {
			delete(he./hostMap, name)
		}
		return valid
	}
	return valid
}

func freePort() (net.Listener, int32, bool) {
	if address, err := net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		var listener net.Listener
		listener, err = net.ListenTCP("tcp", address)
		if err == nil {
			return listener, int32(listener.Addr().(*net.TCPAddr).Port), true
		}
		_ = listener.Close()
	}
	return nil, -1, false
}
