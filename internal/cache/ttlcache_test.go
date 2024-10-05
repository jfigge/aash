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
	assert.Nil(t, cache.config.evictionFn)
	assert.Equal(t, 100, cache.config.reaperBufferSize)
	cache.Close()
}

func TestNewCacheOptionsConfig(t *testing.T) {
	ctx := context.Background()
	cache := NewCache[string, int](ctx,
		OptionDefaultTTL[string, int](time.Hour),
		OptionMaxEntries[string, int](100),
		OptionTouchOnHit[string, int](false),
		OptionReaperInterval[string, int](10*time.Minute),
		OptionAllowEviction[string, int](false),
		OptionAllowReplace[string, int](true),
		OptionEvictionFn[string, int](func(key string, entry int) {}),
		OptionReaperBufferSize[string, int](200),
	)
	assert.Equal(t, time.Hour, cache.config.defaultTTL)
	assert.Equal(t, 100, cache.config.maxEntries)
	assert.Equal(t, false, cache.config.touchOnHit)
	assert.Equal(t, 10*time.Minute, cache.config.reaperInterval)
	assert.Equal(t, false, cache.config.allowEviction)
	assert.Equal(t, true, cache.config.allowReplace)
	assert.NotNil(t, cache.config.evictionFn)
	assert.Equal(t, 200, cache.config.reaperBufferSize)
	cache.Close()
}

func TestNewCacheOptionMaxEntriesRange(t *testing.T) {
	ctx := context.Background()
	cache := NewCache[string, int](ctx,
		OptionMaxEntries[string, int](0),
	)
	assert.Equal(t, 1, cache.config.maxEntries)
	cache.Close()
}

func TestNewCacheOptionReaperIntervalRange(t *testing.T) {
	ctx := context.Background()
	t.Run("too low", func(tt *testing.T) {
		cache := NewCache[string, int](ctx,
			OptionReaperInterval[string, int](time.Second),
		)
		assert.Equal(t, 30*time.Second, cache.config.reaperInterval)
		cache.Close()
	})
	t.Run("too high", func(tt *testing.T) {
		cache := NewCache[string, int](ctx,
			OptionReaperInterval[string, int](7201*time.Second),
		)
		assert.Equal(t, time.Hour, cache.config.reaperInterval)
		cache.Close()
	})
}

func TestNewCacheOptionReaperBufferSizeRange(t *testing.T) {
	ctx := context.Background()
	cache := NewCache[string, int](ctx,
		OptionReaperBufferSize[string, int](-1),
	)
	assert.Equal(t, 1, cache.config.reaperBufferSize)
	cache.Close()
}

func TestNewCacheOptionDefaultTTLRange(t *testing.T) {
	ctx := context.Background()
	cache := NewCache[string, int](ctx,
		OptionDefaultTTL[string, int](30*time.Millisecond),
	)
	assert.Equal(t, time.Minute, cache.config.defaultTTL)
	cache.Close()
}

func TestHappyPath(t *testing.T) {
	evictionCalled := false
	ctx := context.Background()
	c := NewCache[string, int](
		ctx,
		OptionMaxEntries[string, int](2),
		OptionEvictionFn[string, int](func(key string, value int) {
			evictionCalled = key == "A"
		}),
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
	c := NewCache[string, int](
		ctx,
		OptionMaxEntries[string, int](2),
		OptionAllowReplace[string, int](true),
		OptionAllowEviction[string, int](false),
		OptionTouchOnHit[string, int](false),
		OptionEvictionFn[string, int](func(key string, value int) {
			evictionCalled = key == "A"
		}),
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
	c := NewCache[string, int](
		ctx,
		OptionTouchOnHit[string, int](true),
		OptionReaperBufferSize[string, int](1),
		OptionDefaultTTL[string, int](2*time.Second),
		OptionReaperInterval[string, int](100*time.Millisecond),
		OptionEvictionFn[string, int](func(key string, value int) {
			evictedOrder = append(evictedOrder, key)
		}),
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
