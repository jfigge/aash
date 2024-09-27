/*
 * Copyright (C) 2024 by Jason Figge
 */

package cache

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type OptFn func(*config)
type EvictFn func(key interface{}, entry interface{})

type entry[V any] struct {
	item       V
	expiration time.Time
}

type config struct {
	defaultTTL       time.Duration
	maxEntries       int
	evictionFn       EvictFn
	allowEviction    bool
	touchOnHit       bool
	reaperInterval   time.Duration
	reaperBufferSize int
}

type Cache[K comparable, V any] struct {
	parentCtx context.Context
	reaperCtx context.Context
	cancel    context.CancelFunc
	lock      sync.RWMutex
	items     map[K]*entry[V]
	keys      queue[K]
	null      V
	*config
}

func NewCache[K comparable, V any](ctx context.Context, options ...OptFn) *Cache[K, V] {
	reaperCtx, cancel := context.WithCancel(ctx)
	cache := &Cache[K, V]{
		parentCtx: ctx,
		cancel:    cancel,
		reaperCtx: reaperCtx,
		items:     map[K]*entry[V]{},
		config: &config{
			defaultTTL:       30 * time.Minute,
			maxEntries:       50,
			touchOnHit:       true,
			reaperInterval:   5 * time.Minute,
			allowEviction:    true,
			evictionFn:       nil,
			reaperBufferSize: 100,
		},
	}
	for _, option := range options {
		option(cache.config)
	}

	reapChan := make(chan K, cache.reaperBufferSize)
	go cache.reaper(reapChan)
	go cache.reapWorker(reapChan)
	return cache
}

func OptionDefaultTTL(defaultTTL time.Duration) OptFn {
	return func(c *config) {
		c.defaultTTL = checkTTLRange(defaultTTL)
	}
}
func OptionMaxEntries(maxEntries int) OptFn {
	return func(c *config) {
		if maxEntries < 1 {
			maxEntries = 1
		} else if maxEntries > 1000 {
			maxEntries = 1000
		}
		c.maxEntries = maxEntries
	}
}
func OptionEvictionFn(evictionFn EvictFn) OptFn {
	return func(c *config) {
		c.evictionFn = evictionFn
	}
}
func OptionTouchOnHit(touchOnHit bool) OptFn {
	return func(c *config) {
		c.touchOnHit = touchOnHit
	}
}
func OptionReaperInterval(reaperInterval time.Duration) OptFn {
	return func(c *config) {
		if reaperInterval < 30*time.Second {
			reaperInterval = 30 * time.Second
		} else if reaperInterval > time.Hour {
			reaperInterval = time.Hour
		}
		c.reaperInterval = reaperInterval
	}
}
func OptionReaperBufferSize(size int) OptFn {
	return func(c *config) {
		if size < 1 {
			size = 1
		}
		c.reaperBufferSize = size
	}
}
func OptionAllowEviction(allowEviction bool) OptFn {
	return func(c *config) {
		c.allowEviction = allowEviction
	}
}
func checkTTLRange(ttl time.Duration) time.Duration {
	if ttl < time.Minute {
		ttl = time.Minute
	} else if ttl > 24*time.Hour {
		ttl = 24 * time.Hour
	}
	return ttl
}

func (c *Cache[K, V]) Add(key K, value V) bool {
	return c.AddWithTTL(key, value, c.defaultTTL)
}
func (c *Cache[K, V]) AddWithTTL(key K, value V, ttl time.Duration) bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	// Check if key already exists
	if _, ok := c.items[key]; ok {
		return false
	}

	// Check if we hve room and evict if necessary
	if len(c.items) == c.maxEntries {
		if !c.allowEviction {
			return false
		}
		if removeKey, ok := c.keys.Pop(); ok {
			c.evict(removeKey)
		}
	}

	c.items[key] = &entry[V]{
		item:       value,
		expiration: time.Now().Add(checkTTLRange(ttl)),
	}
	c.keys.Push(key)
	return true
}

func (c *Cache[K, V]) Remove(key K) (V, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()
	item, ok := c.items[key]
	if ok {
		delete(c.items, key)
		c.keys.Remove(key)
		return item.item, ok
	}
	return c.null, ok
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	item, ok := c.items[key]
	if !ok {
		return c.null, ok
	}
	if c.touchOnHit {
		c.keys.Remove(key)
		c.keys.Push(key)
		item.expiration = time.Now().Add(c.defaultTTL)
	}
	return item.item, ok
}

func (c *Cache[K, V]) HasKey(key K) bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	_, ok := c.items[key]
	return ok
}

func (c *Cache[K, V]) Close(evictFirst bool) {
	if evictFirst {
		for len(c.items) > 0 {
			if removeKey, ok := c.keys.Pop(); ok {
				c.evict(removeKey)
			}
		}
	}
	c.cancel()
}

func (c *Cache[K, V]) evict(key K) {
	item, ok := c.items[key]
	if !ok {
		return
	}
	defer func() {
		err := recover()
		if err != nil {
			fmt.Printf("recovered from client induced panic: %v", err)
		}
	}()
	delete(c.items, key)
	if c.evictionFn != nil {
		c.evictionFn(key, item)
	}
}

func (c *Cache[K, V]) reaper(reapChan chan<- K) {
	ticker := time.NewTicker(c.reaperInterval)
	defer ticker.Stop()
	for {
		select {
		case <-c.parentCtx.Done():
			return
		case <-c.reaperCtx.Done():
			return
		case <-ticker.C:
			for c.reap(reapChan) {
			}
		}
	}
}

func (c *Cache[K, V]) reap(reapChan chan<- K) bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	for _, key := range c.keys.Items() {
		if c.items[key].expiration.Before(time.Now()) {
			c.keys.Pop()
			select {
			case reapChan <- key:
			default:
				// channel is full. exit reap and signal more to do
				return true
			}
		}
	}
	return false
}

func (c *Cache[K, V]) reapWorker(reapChan <-chan K) {
	for {
		select {
		case <-c.parentCtx.Done():
			return
		case <-c.reaperCtx.Done():
			return
		case key := <-reapChan:
			func() {
				c.lock.Lock()
				defer c.lock.Unlock()
				c.evict(key)
			}()
		}
	}
}
