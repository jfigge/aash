/*
 * Copyright (C) 2024 by Jason Figge
 */

package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"us.figge.auto-ssh/internal/queue"
)

var (
	minReaperInterval = 30 * time.Second
	maxReaperInterval = time.Hour
	minTTLInterval    = time.Minute
)

type OptFn[K comparable, V any] func(*config[K, V])
type EvictFn[K comparable, V any] func(key K, value V)

type entry[V any] struct {
	item       V
	expiration time.Time
}

type config[K comparable, V any] struct {
	defaultTTL       time.Duration
	maxEntries       int
	evictionFn       EvictFn[K, V]
	allowReplace     bool
	allowEviction    bool
	touchOnHit       bool
	reaperInterval   time.Duration
	reaperBufferSize int
}

type Cache[K comparable, V any] struct {
	parentCtx context.Context
	reaperCtx context.Context
	cancel    context.CancelFunc
	deleted   int
	wg        sync.WaitGroup
	lock      sync.RWMutex
	items     map[K]*entry[V]
	keys      *queue.Queue[K]
	null      V
	nullKey   K
	*config[K, V]
}

func NewCache[K comparable, V any](ctx context.Context, options ...OptFn[K, V]) *Cache[K, V] {
	reaperCtx, cancel := context.WithCancel(ctx)
	cache := &Cache[K, V]{
		parentCtx: ctx,
		cancel:    cancel,
		reaperCtx: reaperCtx,
		items:     map[K]*entry[V]{},
		keys:      queue.NewQueue[K](),
		config: &config[K, V]{
			defaultTTL:       30 * time.Minute,
			maxEntries:       -1,
			touchOnHit:       true,
			reaperInterval:   5 * time.Minute,
			allowReplace:     false,
			allowEviction:    true,
			evictionFn:       nil,
			reaperBufferSize: 100,
		},
	}
	for _, option := range options {
		option(cache.config)
	}

	reapChan := make(chan K, cache.reaperBufferSize)
	cache.wg.Add(2)
	go cache.reaper(reapChan)
	go cache.reapWorker(reapChan)
	return cache
}

func OptionDefaultTTL[K comparable, V any](defaultTTL time.Duration) OptFn[K, V] {
	return func(c *config[K, V]) {
		c.defaultTTL = checkTTLRange[K, V](defaultTTL)
	}
}
func OptionMaxEntries[K comparable, V any](maxEntries int) OptFn[K, V] {
	return func(c *config[K, V]) {
		if maxEntries < 1 {
			maxEntries = 1
		}
		c.maxEntries = maxEntries
	}
}
func OptionEvictionFn[K comparable, V any](evictionFn EvictFn[K, V]) OptFn[K, V] {
	return func(c *config[K, V]) {
		c.evictionFn = evictionFn
	}
}
func OptionTouchOnHit[K comparable, V any](touchOnHit bool) OptFn[K, V] {
	return func(c *config[K, V]) {
		c.touchOnHit = touchOnHit
	}
}
func OptionReaperInterval[K comparable, V any](reaperInterval time.Duration) OptFn[K, V] {
	return func(c *config[K, V]) {
		if reaperInterval < minReaperInterval {
			reaperInterval = minReaperInterval
		} else if reaperInterval > maxReaperInterval {
			reaperInterval = maxReaperInterval
		}
		c.reaperInterval = reaperInterval
	}
}
func OptionReaperBufferSize[K comparable, V any](size int) OptFn[K, V] {
	return func(c *config[K, V]) {
		if size < 1 {
			size = 1
		}
		c.reaperBufferSize = size
	}
}
func OptionAllowEviction[K comparable, V any](allowEviction bool) OptFn[K, V] {
	return func(c *config[K, V]) {
		c.allowEviction = allowEviction
	}
}
func OptionAllowReplace[K comparable, V any](allowReplace bool) OptFn[K, V] {
	return func(c *config[K, V]) {
		c.allowReplace = allowReplace
	}
}
func checkTTLRange[K comparable, V any](ttl time.Duration) time.Duration {
	if ttl < minTTLInterval {
		ttl = minTTLInterval
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
		if !c.allowReplace {
			return false
		}
		c.keys.Remove(key)
	} else if len(c.items) == c.maxEntries {
		// ensure we have room for a new item
		if !c.allowEviction {
			return false
		}
		if removeKey, popped := c.keys.Pop(); popped {
			c.evict(removeKey)
		}
	}

	c.items[key] = &entry[V]{
		item:       value,
		expiration: time.Now().Add(checkTTLRange[K, V](ttl)),
	}
	fmt.Printf("%v\n", c.items[key].expiration)
	c.keys.Push(key)
	return true
}

func (c *Cache[K, V]) Entries() int {
	return len(c.items)
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	item, ok := c.items[key]
	if !ok {
		return c.null, false
	}
	if !c.touchOnHit {
		return item.item, true
	}
	c.keys.Remove(key)
	c.keys.Push(key)
	item.expiration = time.Now().Add(c.defaultTTL)
	return item.item, ok
}

func (c *Cache[K, V]) Remove(key K) (V, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()
	item, ok := c.items[key]
	if !ok {
		return c.null, false
	}
	delete(c.items, key)
	c.keys.Remove(key)
	return item.item, ok
}

func (c *Cache[K, V]) HasKey(key K) bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	_, ok := c.items[key]
	return ok
}

func (c *Cache[K, V]) Close() {
	c.cancel()
	c.wg.Wait()
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
		c.evictionFn(key, item.item)
	}
}

func (c *Cache[K, V]) reaper(reapChan chan<- K) {
	defer c.wg.Done()
	ticker := time.NewTicker(c.reaperInterval)
	defer ticker.Stop()
	for {
		select {
		case <-c.parentCtx.Done():
			c.cancel()
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
		if key == c.nullKey {
			continue
		}
		if c.items[key].expiration.Before(time.Now()) {
			c.keys.Pop()
			select {
			case reapChan <- key:
			default:
				// channel is full. exit reap and signal more to do
				return true
			}
		}
		break
	}
	return false
}

func (c *Cache[K, V]) reapWorker(reapChan <-chan K) {
	defer c.wg.Done()
	for {
		select {
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
