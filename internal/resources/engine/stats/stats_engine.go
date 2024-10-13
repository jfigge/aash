/*
 * Copyright (C) 2024 by Jason Figge
 */

package stats

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"us.figge.auto-ssh/internal/core/config"
	engineModels "us.figge.auto-ssh/internal/resources/models"
)

var (
	zeros    = string(make([]byte, 256))
	interval = time.Second * 5
)

type Engine struct {
	lock          sync.Mutex
	statsAddress  string
	statsListener net.Listener
	connections   []net.Conn
	tunnelStats   []*Entry
	updateChan    chan struct{}
	lastUpdate    []byte
	updated       bool
}

func NewEngine() *Engine {
	s := &Engine{
		updateChan: make(chan struct{}),
	}
	return s
}

func (s *Engine) StartStatsTunnel(ctx context.Context, port int) error {
	if config.C.Monitor.StatsPort != -1 {
		var err error
		s.statsAddress = fmt.Sprintf("127.0.0.1:%d", port)
		s.statsListener, err = net.Listen("tcp", s.statsAddress)
		if err != nil {
			fmt.Printf("Warn - Failed to initialize stats monitor: %v\n", err)
			return err
		}
	}
	go s.statsTransmitter(ctx, port)
	return nil
}

func (s *Engine) NewEntry() engineModels.Stats {
	return &Entry{
		statsData:  &statsData{},
		updateChan: s.updateChan,
	}
}

func (s *Engine) statsTransmitter(ctx context.Context, port int) {
	fmt.Printf("  Info  - auto-ssh stats listening on %d\n", port)
	go s.statsBroadcaster(ctx)
	for {
		conn, err := s.statsListener.Accept()
		if err != nil {
			var opErr *net.OpError
			if errors.As(err, &opErr) {
				if opErr.Op == "accept" && opErr.Err.Error() == "use of closed network connection" {
					// CLose quietly and we're likely shutting down
					return
				}
			}
			fmt.Printf("  Error - auto-ssh stats listener accept failed: %v\n", err)
			return
		}
		fmt.Printf("  Info  - Connected stats client\n")
		s.addConnection(conn)
	}
}

func (s *Engine) statsBroadcaster(ctx context.Context) {
	lastBroadcast := time.Now().Add(-interval)
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("  Info  - auto-ssh stats closed\n")
			s.closeAllConnections()
			return
		case <-s.updateChan:
			if !s.updated {
				s.updated = true
				go func() {
					// Don't repeat send data within 5 seconds, but always wait at least 1 second
					// for any pending data to be sent.

					diff := time.Until(lastBroadcast.Add(interval))
					if diff > 0 {
						<-time.NewTimer(diff).C
					} else {
						<-time.NewTimer(time.Second).C
					}
					bs, err := json.Marshal(s.tunnelStats)
					lastBroadcast = time.Now()
					if err == nil {
						s.writeUpdate(bs)
					}
					s.updated = false
				}()
			}
		}
	}
}

func (s *Engine) writeUpdate(update []byte) {
	if !s.lock.TryLock() {
		return
	}
	defer s.lock.Unlock()

	x := 256 - (len(update) % 256)
	s.lastUpdate = append(update, zeros[256-x:]...)
	var alive []net.Conn
	for _, conn := range s.connections {
		if _, err := conn.Write(s.lastUpdate); err != nil {
			fmt.Printf("  Info  - Disconnected stats client\n")
			_ = conn.Close()
		} else {
			alive = append(alive, conn)
		}
	}

	if len(alive) != len(s.connections) {
		s.connections = alive
	}
}

func (s *Engine) addConnection(conn net.Conn) {
	s.lock.Lock()
	defer s.lock.Unlock()
	_, err := conn.Write(s.lastUpdate)
	if err != nil {
		fmt.Printf("  Error - Unable to send current update to new client: %v\n", err)
	}
	s.connections = append(s.connections, conn)
}

func (s *Engine) closeAllConnections() {
	s.lock.Lock()
	defer s.lock.Unlock()
	for _, conn := range s.connections {
		_ = conn.Close()
	}
	_ = s.statsListener.Close()
}
