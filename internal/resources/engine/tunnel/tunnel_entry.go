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

	"us.figge.auto-ssh/internal/core/config"
	engineModels "us.figge.auto-ssh/internal/resources/models"
)

var (
	errInvalidWrite = errors.New("invalid write result")
	connectionIds   = atomic.Int32{}
)

type tunnelData struct {
	*config.Tunnel
	ctx        context.Context
	wg         *sync.WaitGroup
	stats      engineModels.Stats
	lock       *sync.Mutex
	localConns map[int32]*connection
	host       engineModels.HostInternal
	listener   net.Listener
}

type Entry struct {
	//	appCtx context.Context
	*tunnelData
}

func (t *Entry) init(ctx context.Context, stats engineModels.Stats, wg *sync.WaitGroup) {
	t.ctx = ctx
	t.stats = stats
	t.wg = wg
	t.lock = &sync.Mutex{}
	t.localConns = make(map[int32]*connection)
}

func (t *Entry) Start() {
	if t.Status.Running != "Stopped" {
		return
	}
	var err error
	t.Status.Running = "Starting"
	t.listener, err = net.Listen("tcp", t.Local().String())
	if err != nil {
		fmt.Printf("  Error - tunnel (%s) entrance (%s) cannot be created: %v\n", t.Name(), t.Local().String(), err)
		t.Status.Running = "Stopped"
		return
	}
	fmt.Printf("  Info  - tunnel (%s) entrance opened at %s\n", t.Name(), t.Local().String())
	t.wg.Add(1)
	go t.runningAcceptLoop()
	t.Status.Running = "Started"
}

func (t *Entry) Stop() {
	if t.listener != nil {
		t.Status.Running = "Stopping"
		t.listener.Close()
		t.listener = nil
	}
	for _, localConn := range t.localConns {
		localConn.Stop()
	}
}

func (t *Entry) runningAcceptLoop() {
	defer func() {
		t.Status.Running = "Stopped"
		t.wg.Done()
	}()
	for {
		localConn, err := t.listener.Accept()
		if err != nil {
			var opErr *net.OpError
			if errors.As(err, &opErr) && opErr.Op == "accept" && opErr.Err.Error() == "use of closed network connection" {
				fmt.Printf("  Info  - tunnel (%s) closed\n", t.Name())
				return
			}
			fmt.Printf("  Error - tunnel (%s) listener accept failed: %v\n", t.Name(), err)
			return
		}
		fmt.Printf("  Info  - Connected tunnel: %v\n", t.Name())
		go t.forward(localConn)
	}
}

func (t *Entry) forward(localConn net.Conn) {
	var sshConn net.Conn
	if t.host != nil {
		if !t.host.(engineModels.HostInternal).Open() {
			fmt.Printf("  Error - tunnel (%s) unable to open remote connection %s\n", t.Name(), t.Remote())
			return
		}
		var ok bool
		sshConn, ok = t.host.(engineModels.HostInternal).Dial(t.Remote().String())
		if !ok {
			fmt.Printf("  Error - tunnel (%s) unable to dial remote connection %s\n", t.Name(), t.Remote())
			return
		}
	} else {
		// Direct forward
		var err error
		sshConn, err = net.Dial("tcp", t.Remote().String())
		if err != nil {
			fmt.Printf("  Error - tunnel (%s) unable to forward to server %s\n", t.Name(), t.Remote())
			return
		}
	}
	if config.VerboseFlag {
		fmt.Printf("  Info  - tunnel (%s) id:%s connecting to forward server %s\n", t.Name(), t.Id(), t.Remote().String())
	}

	id := connectionIds.Add(1)
	t.localConns[id] = &connection{
		id:        id,
		ctx:       t.ctx,
		name:      t.Name(),
		stats:     t.stats,
		localConn: localConn,
		sshConn:   sshConn,
	}
	defer func() {
		delete(t.localConns, id)
	}()
	t.localConns[id].Start()
}

func (t *Entry) Validate(he engineModels.HostEngineInternal) bool {
	t.tunnelData.Name = strings.TrimSpace(t.tunnelData.Name)
	if t.tunnelData.Name == "" {
		fmt.Printf("  Error - tunnel name cannot be blank\n")
		t.Status.Valid = false
	}
	if t.tunnelData.Remote == nil || t.tunnelData.Remote.IsBlank() {
		fmt.Printf("  Error - tunnel (%s) requires a forward address\n", t.tunnelData.Name)
		t.Status.Valid = false
	} else if !t.tunnelData.Remote.Validate("tunnel", t.tunnelData.Name, "forward address", true, false) {
		t.Status.Valid = false
	}

	if (t.tunnelData.Local == nil || t.tunnelData.Local.IsBlank()) && t.tunnelData.Remote != nil && t.tunnelData.Remote.IsValid() {
		fmt.Printf("  Warn  - tunnel (%s) Local entrance undefined. Defaulting to 127.0.0.1:%d\n", t.tunnelData.Name, t.tunnelData.Remote.Port())
		t.tunnelData.Local = config.NewAddress(fmt.Sprintf("127.0.0.1:%d", t.tunnelData.Remote.Port()))
	}
	if t.tunnelData.Local == nil || t.tunnelData.Local.IsBlank() {
		fmt.Printf("  Error - tunnel (%s) missing a local address that cannot be derived\n", t.tunnelData.Name)
	} else if !t.tunnelData.Local.Validate("tunnel", t.tunnelData.Name, "local address", true, false) {
		t.Status.Valid = false
	}

	t.tunnelData.Host = strings.TrimSpace(t.tunnelData.Host)
	if t.tunnelData.Host == "" {
		fmt.Printf("  Info  - tunnel (%s) exits on the local host\n", t.tunnelData.Name)
	} else if host, ok := he.Host(t.tunnelData.Host); !ok {
		fmt.Printf("  Error - tunnel (%s) remote host (%s) undefined\n", t.tunnelData.Name, t.tunnelData.Host)
		t.Status.Valid = false
	} else if !host.Valid() {
		fmt.Printf("  Error - tunnel (%s) remote host (%s) is invalid\n", t.tunnelData.Name, t.tunnelData.Host)
		t.Status.Valid = false
	} else if t.Status.Valid {
		t.host = host.(engineModels.HostInternal)
		t.host.Referenced()
	}

	if config.VerboseFlag && t.Status.Valid {
		fmt.Printf("  Info  - tunnel (%s) validated\n", t.tunnelData.Name)
	}

	//t.stats = &TunnelStats{
	//	Name: t.Name,
	//	Port: t.Local.Port(),
	//}

	return t.Status.Valid
}
func (t *Entry) Id() string {
	return t.tunnelData.Id
}
func (t *Entry) Name() string {
	return t.tunnelData.Name
}
func (t *Entry) Local() *config.Address {
	return t.tunnelData.Local
}
func (t *Entry) Remote() *config.Address {
	return t.tunnelData.Remote
}
func (t *Entry) Host() string {
	return t.tunnelData.Host
}
func (t *Entry) Valid() bool {
	return t.tunnelData.Status.Valid
}
func (t *Entry) Running() string {
	return t.tunnelData.Status.Running
}
func (t *Entry) Metadata() *config.Metadata {
	return t.tunnelData.Metadata
}
