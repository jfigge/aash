/*
 * Copyright (C) 2024 by Jason Figge
 */

package stats

import (
	"fmt"
	"sync/atomic"
	"time"
)

var (
	currentConnections = atomic.Int32{}
	totalConnections   = atomic.Int32{}
)

type statsData struct {
	Id          int       `json:"i" title:"Id"   format:"%%%ds "  sort:"%[2]s%[1]s"`
	Name        string    `json:"n" title:"Name" format:"%%-%ds " sort:"%[1]s%[2]s"`
	Port        int       `json:"p" title:"Port" format:"%%%ds "  sort:"%[2]s%[1]s"`
	In          int64     `json:"r" title:"Rcvd" format:"%%%ds "  sort:"%[2]s%[1]s"`
	Out         int64     `json:"t" title:"Sent" format:"%%%ds "  sort:"%[2]s%[1]s"`
	Connected   int       `json:"o" title:"Open" format:"%%%ds "  sort:"%[2]s%[1]s"`
	Connections int       `json:"c" title:"Used" format:"%%%ds "  sort:"%[2]s%[1]s"`
	JumpTunnel  bool      `json:"j" title:"Jump" format:"%%%ds "  sort:"%[2]s%[1]s"`
	LastUpdate  time.Time `json:"u" title:"Last" format:"%%-%ds " sort:"%[1]s%[2]s"`
}

type Entry struct {
	*statsData
	updateChan chan struct{}
}

func (e Entry) Connected() int {
	currentConnections.Add(1)
	totalConnections.Add(1)
	return int(currentConnections.Load())
}

func (e Entry) Disconnected() {
	currentConnections.Add(-1)
}

func (e Entry) Received(n int64) {
	fmt.Printf("  Info  - Recieved %d\n", n)
	e.In += n
}

func (e Entry) Transmitted(n int64) {
	fmt.Printf("  Info  - Transmitted %d\n", n)
	e.Out += n
}

func (e Entry) Updated() {
	e.LastUpdate = time.Now()

}
