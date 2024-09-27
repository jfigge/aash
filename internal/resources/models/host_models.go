/*
 * Copyright (C) 2024 by Jason Figge
 */

package models

import (
	"us.figge.auto-ssh/internal/core/config"
)

type Host interface {
	Hosts() []*config.Host
}
