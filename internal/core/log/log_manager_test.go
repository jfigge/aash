/*
 * Copyright (C) 2024 by Jason Figge
 */

package log

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLogManager_expireMessages(t *testing.T) {
	past := time.Now().Add(-time.Hour)
	future := time.Now().Add(+time.Hour)
	tests := map[string]struct {
		history []*msgEntry
		actual  []string
	}{
		"empty": {
			history: []*msgEntry{},
			actual:  []string{},
		},
		"expire-0": {
			history: []*msgEntry{
				{expiration: future, msg: "1"},
				{expiration: future, msg: "2"},
				{expiration: future, msg: "3"},
			},
			actual: []string{"1", "2", "3"},
		},
		"expire-1": {
			history: []*msgEntry{
				{expiration: past, msg: "1"},
				{expiration: future, msg: "2"},
				{expiration: future, msg: "3"},
			},
			actual: []string{"2", "3"},
		},
		"expire-2": {
			history: []*msgEntry{
				{expiration: past, msg: "1"},
				{expiration: past, msg: "2"},
				{expiration: future, msg: "3"},
			},
			actual: []string{"3"},
		},
		"expire-all": {
			history: []*msgEntry{
				{expiration: past, msg: "1"},
				{expiration: past, msg: "2"},
				{expiration: past, msg: "3"},
			},
			actual: []string{},
		},
	}
	for name, test := range tests {
		t.Run(name, func(tt *testing.T) {
			defaultLM.history = test.history
			defaultLM.expireMessages()
			assert.Equal(tt, strings.Join(test.actual, ","), strings.Join(Messages(), ","))
		})
	}
}
