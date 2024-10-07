/*
 * Copyright (C) 2024 by Jason Figge
 */

package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"us.figge.auto-ssh/internal/core/config"
)

var (
	zeros    = string(make([]byte, 256))
	interval = time.Second * 5
)

type TunnelStats struct {
	Id          int       `json:"i" title:"Id"   format:"%%%ds "  sort:"%[2]s%[1]s"`
	Name        string    `json:"n" title:"Name" format:"%%-%ds " sort:"%[1]s%[2]s"`
	Port        int       `json:"p" title:"Port" format:"%%%ds "  sort:"%[2]s%[1]s"`
	Received    int64     `json:"r" title:"Rcvd" format:"%%%ds "  sort:"%[2]s%[1]s"`
	Transmitted int64     `json:"t" title:"Sent" format:"%%%ds "  sort:"%[2]s%[1]s"`
	Connected   int       `json:"o" title:"Open" format:"%%%ds "  sort:"%[2]s%[1]s"`
	Connections int       `json:"c" title:"Used" format:"%%%ds "  sort:"%[2]s%[1]s"`
	JumpTunnel  bool      `json:"j" title:"Jump" format:"%%%ds "  sort:"%[2]s%[1]s"`
	Updated     time.Time `json:"u" title:"Last" format:"%%-%ds " sort:"%[1]s%[2]s"`
	// private properties must be listed last
}

type StatsEngine struct {
	lock          sync.Mutex
	statsAddress  string
	statsListener net.Listener
	connections   []net.Conn
	tunnelStats   []*TunnelStats
	updateChan    chan struct{}
	lastUpdate    []byte
	updated       bool
}

func NewStatsEngine() *StatsEngine {
	s := &StatsEngine{
		updateChan: make(chan struct{}),
	}
	return s
}

func (s *StatsEngine) StartStatsTunnel(ctx context.Context, port int) error {
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

func (s *StatsEngine) statsTransmitter(ctx context.Context, port int) {
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

func (s *StatsEngine) statsBroadcaster(ctx context.Context) {
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

func (s *StatsEngine) writeUpdate(update []byte) {
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

func (s *StatsEngine) addConnection(conn net.Conn) {
	s.lock.Lock()
	defer s.lock.Unlock()
	_, err := conn.Write(s.lastUpdate)
	if err != nil {
		fmt.Printf("  Error - Unable to send current update to new client: %v\n", err)
	}
	s.connections = append(s.connections, conn)
}

func (s *StatsEngine) closeAllConnections() {
	s.lock.Lock()
	defer s.lock.Unlock()
	for _, conn := range s.connections {
		_ = conn.Close()
	}
	_ = s.statsListener.Close()
}
