/*
 * Copyright (C) 2024 by Jason Figge
 */

package tunnel

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"us.figge.auto-ssh/internal/core/config"
	engineModels "us.figge.auto-ssh/internal/resources/models"
)

type tunnelConn struct {
	id        string
	name      string
	stats     engineModels.Stats
	conns     [2]net.Conn
	connected [2]bool
}

func NewTunnelConnection(name string, id string, stats engineModels.Stats, sshConn net.Conn, localConn net.Conn) *tunnelConn {
	return &tunnelConn{
		name:      name,
		id:        id,
		stats:     stats,
		conns:     [2]net.Conn{localConn, sshConn},
		connected: [2]bool{true, true},
	}
}

func (t *tunnelConn) Start(ctx context.Context) {
	tunnelCtx, cancel := context.WithCancel(ctx)
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		t.send(tunnelCtx, 0, "send")
		wg.Done()
	}()
	go func() {
		t.send(tunnelCtx, 1, "receive")
		wg.Done()
	}()
	wg.Wait()
	cancel()
	if config.VerboseFlag {
		fmt.Printf("  Info  - id:%s closing connection %s\n", t.id, t.conns[0].RemoteAddr())
	}
}

func (t *tunnelConn) send(ctx context.Context, index int, name string) {
	if config.VerboseFlag {
		fmt.Printf("  Info  - tunnel (%s) id:%s %s tunnel opened\n", t.name, t.id, name)
	}
	err := t.copy(t.conns[index], t.conns[1-index], index == 0)
	if err != nil && config.VerboseFlag {
		fmt.Printf("  Error - tunnel (%s) id:%s encountered a closed tunnel: %v\n", t.name, t.id, err)
	}
	t.connected[index] = false
	if config.VerboseFlag {
		fmt.Printf("  Info  - tunnel (%s) id:%s %s tunnel closed\n", t.name, t.id, name)
	}
	if t.connected[1-index] {
		go t.autoClose(ctx)
	}
}

func (t *tunnelConn) copy(src io.Reader, dst io.Writer, read bool) (err error) {
	buf := make([]byte, 32*1024)
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			fmt.Printf("%v => %s\n", read, buf[0:nr])
			nw, ew := dst.Write(buf[0:nr])
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = errInvalidWrite
				}
			}
			if read {
				t.stats.Received(int64(nw))
			} else {
				t.stats.Transmitted(int64(nw))
			}
			t.stats.Updated()

			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return err
}

func (t *tunnelConn) autoClose(ctx context.Context) {
	status := "terminated"
	if config.VerboseFlag {
		fmt.Printf("  Info  - tunnel (%s) id:%s auto-closer initiated\n", t.name, t.id)
	}
	timer := time.NewTimer(30 * time.Second)
	select {
	case <-timer.C:
		status = "triggered"
	case <-ctx.Done():
	}
	for i := range 2 {
		if t.conns[i] != nil {
			_ = t.conns[i].Close()
		}
	}
	if config.VerboseFlag {
		fmt.Printf("  Info  - tunnel (%s) id:%s auto-closer %s\n", t.name, t.id, status)
	}
}
