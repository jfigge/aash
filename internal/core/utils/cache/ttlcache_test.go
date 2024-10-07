/*
 * Copyright (C) 2024 by Jason Figge
 */

package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewCacheDefaultConfig(t *testing.T) {
	ctx := context.Background()
	cache := NewCache[string, int](ctx)
	assert.Equal(t, 30*time.Minute, cache.config.defaultTTL)
	assert.Equal(t, -1, cache.config.maxEntries)
	assert.Equal(t, true, cache.config.touchOnHit)
	assert.Equal(t, 5*time.Minute, cache.config.reaperInterval)
	assert.Equal(t, true, cache.config.allowEviction)
	assert.Equal(t, false, cache.config.allowReplace)
	assert.Equal(t, 100, cache.config.reaperBufferSize)
	cache.Close()
}

func TestNewCacheOptionsConfig(t *testing.T) {
	ctx := context.Background()
	cache := NewCache[string, int](ctx,
		OptionDefaultTTL(time.Hour),
		OptionMaxEntries(100),
		OptionTouchOnHit(false),
		OptionReaperInterval(10*time.Minute),
		OptionAllowEviction(false),
		OptionAllowReplace(true),
		OptionReaperBufferSize(200),
	)
	assert.Equal(t, time.Hour, cache.config.defaultTTL)
	assert.Equal(t, 100, cache.config.maxEntries)
	assert.Equal(t, false, cache.config.touchOnHit)
	assert.Equal(t, 10*time.Minute, cache.config.reaperInterval)
	assert.Equal(t, false, cache.config.allowEviction)
	assert.Equal(t, true, cache.config.allowReplace)
	assert.Equal(t, 200, cache.config.reaperBufferSize)
	cache.Close()
}

func TestNewCacheOptionMaxEntriesRange(t *testing.T) {
	ctx := context.Background()
	cache := NewCache[string, int](ctx,
		OptionMaxEntries(0),
	)
	assert.Equal(t, 1, cache.config.maxEntries)
	cache.Close()
}

func TestNewCacheOptionReaperIntervalRange(t *testing.T) {
	ctx := context.Background()
	t.Run("too low", func(tt *testing.T) {
		cache := NewCache[string, int](ctx,
			OptionReaperInterval(time.Second),
		)
		assert.Equal(t, 30*time.Second, cache.config.reaperInterval)
		cache.Close()
	})
	t.Run("too high", func(tt *testing.T) {
		cache := NewCache[string, int](ctx,
			OptionReaperInterval(7201*time.Second),
		)
		assert.Equal(t, time.Hour, cache.config.reaperInterval)
		cache.Close()
	})
}

func TestNewCacheOptionReaperBufferSizeRange(t *testing.T) {
	ctx := context.Background()
	cache := NewCache[string, int](ctx,
		OptionReaperBufferSize(-1),
	)
	assert.Equal(t, 1, cache.config.reaperBufferSize)
	cache.Close()
}

func TestNewCacheOptionDefaultTTLRange(t *testing.T) {
	ctx := context.Background()
	cache := NewCache[string, int](ctx,
		OptionDefaultTTL(30*time.Millisecond),
	)
	assert.Equal(t, time.Minute, cache.config.defaultTTL)
	cache.Close()
}

func TestHappyPath(t *testing.T) {
	evictionCalled := false
	ctx := context.Background()
	c := NewCacheWithEvict[string, int](
		ctx,
		func(key string, value int) {
			evictionCalled = key == "A"
		},
		OptionMaxEntries(2),
	)

	// Adds
	assert.False(t, c.HasKey("A"))
	assert.True(t, c.Add("A", 1))
	assert.True(t, c.HasKey("A"))
	assert.False(t, c.Add("A", 1))
	assert.True(t, c.HasKey("A"))
	assert.True(t, c.Add("B", 2))
	assert.False(t, evictionCalled)
	assert.True(t, c.Add("C", 3))
	assert.True(t, evictionCalled)
	assert.Equal(t, 2, c.Entries())

	// Gets
	value, ok := c.Get("A")
	assert.False(t, ok)
	assert.Equal(t, value, 0)
	value, ok = c.Get("B")
	assert.True(t, ok)
	assert.Equal(t, value, 2)
	value, ok = c.Get("C")
	assert.True(t, ok)
	assert.Equal(t, value, 3)

	// Removes
	value, ok = c.Remove("A")
	assert.False(t, ok)
	assert.Equal(t, value, 0)
	value, ok = c.Remove("B")
	assert.True(t, ok)
	assert.Equal(t, value, 2)
	value, ok = c.Remove("C")
	assert.True(t, ok)
	assert.Equal(t, value, 3)
	assert.Equal(t, 0, c.Entries())

	c.Close()
}

func TestAlternativeHappyPath(t *testing.T) {
	evictionCalled := false
	ctx := context.Background()
	c := NewCacheWithEvict[string, int](
		ctx,
		func(key string, value int) {
			evictionCalled = key == "A"
		},
		OptionMaxEntries(2),
		OptionAllowReplace(true),
		OptionAllowEviction(false),
		OptionTouchOnHit(false),
	)

	// Adds
	assert.False(t, c.HasKey("A"))
	assert.True(t, c.Add("A", 1))
	assert.True(t, c.HasKey("A"))
	assert.True(t, c.Add("A", 11))
	assert.True(t, c.HasKey("A"))
	assert.True(t, c.Add("B", 2))
	assert.False(t, evictionCalled)
	assert.False(t, c.Add("C", 3))
	assert.False(t, evictionCalled)
	assert.Equal(t, 2, c.Entries())

	// Removes
	value, ok := c.Get("A")
	assert.True(t, ok)
	assert.Equal(t, value, 11)

	// Removes
	value, ok = c.Remove("A")
	assert.True(t, ok)
	assert.Equal(t, value, 11)
	value, ok = c.Remove("B")
	assert.True(t, ok)
	assert.Equal(t, value, 2)
	value, ok = c.Remove("C")
	assert.False(t, ok)
	assert.Equal(t, value, 0)
	assert.Equal(t, 0, c.Entries())

	c.Close()
}

func TestEviction(t *testing.T) {
	ctx := context.Background()
	minReaperInterval = 1
	minTTLInterval = 1
	var evictedOrder []string
	c := NewCacheWithEvict[string, int](
		ctx,
		func(key string, value int) {
			evictedOrder = append(evictedOrder, key)
		},
		OptionTouchOnHit(true),
		OptionReaperBufferSize(1),
		OptionDefaultTTL(2*time.Second),
		OptionReaperInterval(100*time.Millisecond),
	)

	c.Add("A", 1)
	c.Add("B", 2)
	c.Add("C", 3)
	c.Add("D", 4)
	time.Sleep(1 * time.Second)
	value, ok := c.Get("A")
	assert.True(t, ok)
	assert.Equal(t, 1, value)
	time.Sleep(3 * time.Second)
	assert.Equal(t, []string{"B", "C", "D", "A"}, evictedOrder)
	c.Close()
}
