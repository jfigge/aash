/*
 * Copyright (C) 2024 by Jason Figge
 */

package models

import (
	"context"
)

type StatsEngine interface {
	StartStatsTunnel(ctx context.Context, port int) error
	NewEntry() Stats
}

type Stats interface {
	Connected() int
	Disconnected()
	Received(i int64)
	Transmitted(i int64)
	Updated()
}
