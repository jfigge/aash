/*
 * Copyright (C) 2024 by Jason Figge
 */

package log

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type msgEntry struct {
	expiration time.Time
	msg        string
}

type LogManager struct {
	history []*msgEntry
	lock    sync.Mutex
	stdChn  chan string
	ctx     context.Context
	size    int
	ttl     time.Duration
}

var (
	defaultLM *LogManager
)

type LogOption func(lm *LogManager)

func LogOptionSize(size int) LogOption {
	return func(lm *LogManager) {
		lm.size = size
	}
}

func LogOptionTTL(ttl time.Duration) LogOption {
	return func(lm *LogManager) {
		lm.ttl = ttl
	}
}

func init() {
	defaultLM = &LogManager{
		history: make([]*msgEntry, 0),
		stdChn:  make(chan string, 100),
		lock:    sync.Mutex{},
		size:    1000,
		ttl:     time.Hour * 24,
	}
}

func InitLogManager(options ...LogOption) {
	for _, option := range options {
		option(defaultLM)
	}
}

func Start(ctx context.Context) {
	if defaultLM.ctx == nil {
		defaultLM.ctx = ctx
		go defaultLM.captureMessages()
		go defaultLM.sweeper()
	}
}

func (lm *LogManager) sweeper() {
	for {
		t := time.NewTicker(lm.ttl)
		select {
		case <-lm.ctx.Done():
			t.Stop()
			return
		case <-t.C:
			lm.expireMessages()
		}
	}
}

func (lm *LogManager) expireMessages() {
	lm.lock.Lock()
	defer lm.lock.Unlock()

	expired := false
	index := 0
	indexLast := len(lm.history) - 1
	for ; index <= indexLast && lm.history[index].expiration.Before(time.Now()); index++ {
		expired = true
	}
	if expired {
		lm.history = lm.history[index:]
	}
}

func (lm *LogManager) captureMessages() {
	for {
		select {
		case <-lm.ctx.Done():
			return
		case msg := <-lm.stdChn:
			func() {
				lm.lock.Lock()
				defer lm.lock.Unlock()
				lm.history = append(lm.history, &msgEntry{expiration: time.Now().Add(lm.ttl), msg: msg})
				fmt.Printf(msg)
				if len(lm.history) > lm.size {
					lm.history = lm.history[:lm.size]
				}
			}()
		}
	}
}

func Printf(format string, v ...any) {
	defaultLM.stdChn <- fmt.Sprintf(format, v...)
}

func Messages() []string {
	defaultLM.lock.Lock()
	defer defaultLM.lock.Unlock()
	messages := make([]string, len(defaultLM.history))
	for i, entry := range defaultLM.history {
		messages[i] = entry.msg
	}
	return messages
}
