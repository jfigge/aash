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
	"us.figge.auto-ssh/internal/resources/models"
)

type connection struct {
	id        int32
	name      string
	ctx       context.Context
	stats     models.Stats
	localConn net.Conn
	sshConn   net.Conn
}

func (c *connection) Start() {
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		c.transcieve(c.localConn, c.sshConn, true, "local => remove")
		c.autoClose()
		wg.Done()
	}()
	go func() {
		c.transcieve(c.sshConn, c.localConn, false, "remove => local")
		c.autoClose()
		wg.Done()
	}()
	wg.Wait()
}

func (c *connection) transcieve(src net.Conn, dest net.Conn, read bool, name string) {
	if config.VerboseFlag {
		fmt.Printf("  Info  - tunnel (%s) id:%d started %s\n", c.name, c.id, name)
	}
	c.copy(src, dest, read)
	if config.VerboseFlag {
		fmt.Printf("  Info  - tunnel (%s) id:%d stopped %s\n", c.name, c.id, name)
	}
}

func (c *connection) copy(src io.Reader, dst io.Writer, read bool) (err error) {
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
				c.stats.Received(int64(nw))
			} else {
				c.stats.Transmitted(int64(nw))
			}
			c.stats.Updated()

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

func (c *connection) Stop() {
	c.autoClose()
}

func (c *connection) autoClose() {
	status := "terminated"
	if config.VerboseFlag {
		fmt.Printf("  Info  - tunnel (%s) id:%d auto-closer initiated\n", c.name, c.id)
	}
	timer := time.NewTimer(30 * time.Second)
	select {
	case <-timer.C:
		status = "triggered"
	case <-c.ctx.Done():
	}
	if c.sshConn != nil {
		_ = c.sshConn.Close()
	}
	if c.localConn != nil {
		_ = c.localConn.Close()
	}
	if config.VerboseFlag {
		fmt.Printf("  Info  - tunnel (%s) id:%s auto-closer %s\n", c.name, c.id, status)
	}
}
