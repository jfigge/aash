/*
 * Copyright (C) 2024 by Jason Figge
 */

package tunnel

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"us.figge.auto-ssh/internal/core/config"
	"us.figge.auto-ssh/internal/resources/engine/stats"
	engineModels "us.figge.auto-ssh/internal/resources/models"
)

var (
	errInvalidWrite = errors.New("invalid write result")
	connection      = atomic.Int32{}
	connections     = atomic.Int32{}
)

type tunnelData struct {
	*config.Tunnel
	lock    sync.Mutex
	host    engineModels.HostInternal
	conns   []net.Conn
	stats   *stats.TunnelStats
	cancel  context.CancelFunc
	wg      *sync.WaitGroup
	valid   bool
	running string
}

type TunnelEntry struct {
	tunnelData
}

func (t *TunnelEntry) init(wg *sync.WaitGroup) {
	t.wg = wg
}

func (t *TunnelEntry) Start(ctx context.Context) {
	if t.running != "Stopped" {
		return
	}
	t.wg.Add(1)
	go func() {
		defer t.wg.Done()
		t.open(ctx)
	}()
}

func (t *TunnelEntry) Stop(ctx context.Context) {
	if t.cancel != nil {
		t.running = "Stopping"
		t.cancel()
	}
}

func (t *TunnelEntry) open(ctx context.Context) {
	t.running = "Starting"
	defer func() {
		t.running = "Stopped"
	}()
	ctx, t.cancel = context.WithCancel(ctx)
	localListener, err := net.Listen("tcp", t.Local().String())
	if err != nil {
		fmt.Printf("  Error - tunnel (%s) entrance (%s) cannot be created: %v\n", t.Name(), t.Local().String(), err)
		return
	}
	fmt.Printf("  Info  - tunnel (%s) entrance opened at %s\n", t.Name(), t.Local().String())
	go t.waitForTermination(ctx, localListener)
	t.running = "Started"
	for {
		var localConn net.Conn
		localConn, err = localListener.Accept()
		if err != nil {
			var opErr *net.OpError
			if errors.As(err, &opErr) && opErr.Op == "accept" && opErr.Err.Error() == "use of closed network connection" {
				// Close quietly and we're likely shutting down
				return
			}
			fmt.Printf("  Error - tunnel (%s) listener accept failed: %v\n", t.Name(), err)
			return
		}
		t.stats.Updated = time.Now()
		//t.stats.updateChan <- struct{}{}
		fmt.Printf("  Info  - Connected tunnel: %v\n", t.Name())
		go t.forward(ctx, localConn)
	}
}

func (t *TunnelEntry) forward(ctx context.Context, localConn net.Conn) {
	t.addConnection(localConn)
	defer t.removeConnection(localConn)
	connection.Add(1)
	id := connection.Load()

	if config.VerboseFlag {
		fmt.Printf("  Info  - tunnel (%s) id:%d conneting to forward server %s\n", t.Name(), id, t.Remote().String())
	}

	var sshConn net.Conn
	if t.host != nil {
		if !t.host.(engineModels.HostInternal).Open() {
			// TODO Failed to connect
			return
		}
		var ok bool
		sshConn, ok = t.host.(engineModels.HostInternal).Dial(t.Remote().String())
		if !ok {
			// TODO failed to connect
			return
		}
	} else {
		// Direct forward
		var err error
		sshConn, err = net.Dial("tcp", t.Remote().String())
		if err != nil {
			fmt.Printf("  Error - tunnel (%s) id:%d unable to forward to server %s\n", t.Name(), id, t.Remote().String())
			return
		}
	}

	NewTunnelConnection(t.Name(), int(id), sshConn, localConn).Start(ctx)
}

func (t *TunnelEntry) Validate(he engineModels.HostEngineInternal) bool {
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
	} else if !host.Valid() {
		fmt.Printf("  Error - tunnel (%s) remote host (%s) is invalid\n", t.tunnelData.Name, t.tunnelData.Host)
		t.valid = false
	} else if t.valid {
		t.host = host.(engineModels.HostInternal)
		t.host.Referenced()
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
func (t *TunnelEntry) Running() string {
	return t.tunnelData.running
}

func (t *TunnelEntry) Metadata() *config.Metadata {
	return t.tunnelData.Metadata
}

func (t *TunnelEntry) waitForTermination(ctx context.Context, localListener net.Listener) {
	<-ctx.Done()
	fmt.Printf("  Info  - tunnel (%s) stopped listening on %s\n", t.Name(), t.Local().String())
	_ = localListener.Close()
	t.lock.Lock()
	defer t.lock.Unlock()
	for _, conn := range t.conns {
		_ = conn.Close()
	}
	t.conns = []net.Conn{}
	t.cancel = nil
}

func (t *TunnelEntry) addConnection(conn net.Conn) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.stats.Connected++
	t.stats.Connections++
	t.conns = append(t.conns, conn)
}

func (t *TunnelEntry) removeConnection(conn net.Conn) {
	t.lock.Lock()
	defer t.lock.Unlock()
	conns := make([]net.Conn, 0, len(t.conns)-1)
	for _, c := range t.conns {
		if conn != c {
			conns = append(conns, c)
		}
	}
	_ = conn.Close()
	t.stats.Connected--
	t.conns = conns
}
