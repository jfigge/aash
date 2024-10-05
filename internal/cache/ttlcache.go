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

type OptFn func(*config)
type EvictFn[K comparable, V any] func(key K, value V)

type entry[V any] struct {
	item       V
	expiration time.Time
}

type config struct {
	defaultTTL       time.Duration
	maxEntries       int
	allowReplace     bool
	allowEviction    bool
	touchOnHit       bool
	reaperInterval   time.Duration
	reaperBufferSize int
}

type Cache[K comparable, V any] struct {
	parentCtx  context.Context
	reaperCtx  context.Context
	cancel     context.CancelFunc
	deleted    int
	wg         sync.WaitGroup
	lock       sync.RWMutex
	items      map[K]*entry[V]
	keys       *queue.Queue[K]
	null       V
	nullKey    K
	evictionFn EvictFn[K, V]
	*config
}

func NewCache[K comparable, V any](ctx context.Context, options ...OptFn) *Cache[K, V] {
	return NewCacheWithEvict(ctx, func(K, V) {}, options...)
}
func NewCacheWithEvict[K comparable, V any](ctx context.Context, evictFn EvictFn[K, V], options ...OptFn) *Cache[K, V] {
	reaperCtx, cancel := context.WithCancel(ctx)
	cache := &Cache[K, V]{
		parentCtx:  ctx,
		cancel:     cancel,
		reaperCtx:  reaperCtx,
		items:      map[K]*entry[V]{},
		keys:       queue.NewQueue[K](),
		evictionFn: evictFn,
		config: &config{
			defaultTTL:       30 * time.Minute,
			maxEntries:       -1,
			touchOnHit:       true,
			reaperInterval:   5 * time.Minute,
			allowReplace:     false,
			allowEviction:    true,
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

func OptionDefaultTTL(defaultTTL time.Duration) OptFn {
	return func(c *config) {
		c.defaultTTL = checkTTLRange(defaultTTL)
	}
}
func OptionMaxEntries(maxEntries int) OptFn {
	return func(c *config) {
		if maxEntries < 1 {
			maxEntries = 1
		}
		c.maxEntries = maxEntries
	}
}
func OptionTouchOnHit(touchOnHit bool) OptFn {
	return func(c *config) {
		c.touchOnHit = touchOnHit
	}
}
func OptionReaperInterval(reaperInterval time.Duration) OptFn {
	return func(c *config) {
		if reaperInterval < minReaperInterval {
			reaperInterval = minReaperInterval
		} else if reaperInterval > maxReaperInterval {
			reaperInterval = maxReaperInterval
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
func OptionAllowReplace(allowReplace bool) OptFn {
	return func(c *config) {
		c.allowReplace = allowReplace
	}
}
func checkTTLRange(ttl time.Duration) time.Duration {
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
		expiration: time.Now().Add(checkTTLRange(ttl)),
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
