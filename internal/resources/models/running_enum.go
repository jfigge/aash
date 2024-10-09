/*
 * Copyright (C) 2024 by Jason Figge
 */

package models

type Running int

var (
	entries = [...]string{"Stopped", "Starting", "Started", "Stopping"}
)

const (
	Stopped Running = iota
	Starting
	Started
	Stopping
)

func RunningEnums() [4]string {
	return entries

}
func (r Running) String() string {
	return entries[r]
}

func (r Running) Index() int {
	return int(r)
}
