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
)

type tunnelConn struct {
	id    int
	name  string
	conns [2]net.Conn
	ctxs  [2]context.Context
}

func NewTunnelConnection(name string, id int, sshConn net.Conn, localConn net.Conn) *tunnelConn {
	return &tunnelConn{
		id:    id,
		name:  name,
		conns: [2]net.Conn{localConn, sshConn},
		ctxs:  [2]context.Context{context.Background(), context.Background()},
	}
}

func (t *tunnelConn) Start(ctx context.Context) {
	tunnelCtx, cancel := context.WithCancel(ctx)
	closer := func() {
		t.autoClose(tunnelCtx)
	}
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go t.send(wg, 0, closer, "send")
	go t.send(wg, 1, closer, "receive")
	wg.Wait()
	cancel()
	if config.VerboseFlag {
		fmt.Printf("  Info  - id:%d c:%d closing connection %s\n", t.id, connections.Load(), t.conns[0].RemoteAddr())
	}
}

func (t *tunnelConn) send(wg *sync.WaitGroup, index int, closer func(), name string) {
	connections.Add(1)
	id1 := connections.Load()
	if config.VerboseFlag {
		fmt.Printf("  Info  - tunnel (%s) id:%d c:%d %s tunnel opened\n", t.name, t.id, id1, name)
	}
	defer func() {
		wg.Done()
		t.ctxs[index].Done()
	}()
	err1 := t.copy(t.conns[index], t.conns[1-index], true)
	if err1 != nil && config.VerboseFlag {
		fmt.Printf("  Error - tunnel (%s) %s encountered a closed tunnel: %v\n", t.name, name, err1)
	}
	connections.Add(-1)
	id1 = connections.Load()
	if config.VerboseFlag {
		fmt.Printf("  Info  - tunnel (%s) id:%d c:%d %s tunnel closed\n", t.name, t.id, id1, name)
	}
	if t.ctxs[1-index].Err() == nil {
		go closer()
	}
}

func (t *tunnelConn) copy(dst io.Writer, src io.Reader, read bool) (err error) {
	buf := make([]byte, 32*1024)
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = errInvalidWrite
				}
			}
			//if read {
			//	t.stats.Received += int64(nw)
			//} else {
			//	t.stats.Transmitted += int64(nw)
			//}
			//t.stats.Updated = time.Now()
			//t.stats.updateChan <- struct{}{}

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
		fmt.Printf("  Info  - tunnel (%s) id:%d c:%d auto-closer initiated\n", t.name, t.id, connections.Load())
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
			_ = t.ctxs[i].Done()
		}
	}
	if config.VerboseFlag {
		fmt.Printf("  Info  - tunnel (%s) id:%d c:%d auto-closer %s\n", t.name, t.id, connections.Load(), status)
	}
}
