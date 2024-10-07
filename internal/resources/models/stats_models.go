/*
 * Copyright (C) 2024 by Jason Figge
 */

package models

import (
	"context"
)

type Stats interface {
	StartStatsTunnel(ctx context.Context, port int) error
}
