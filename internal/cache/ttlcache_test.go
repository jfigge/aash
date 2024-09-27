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
	assert.Equal(t, 50, cache.config.maxEntries)
	assert.Equal(t, true, cache.config.touchOnHit)
	assert.Equal(t, 5*time.Minute, cache.config.reaperInterval)
	assert.Equal(t, true, cache.config.allowEviction)
	assert.Nil(t, cache.config.evictionFn)
	assert.Equal(t, 100, cache.config.reaperBufferSize)
	cache.Close(false)
}

func TestNewCacheOptionsConfig(t *testing.T) {
	ctx := context.Background()
	cache := NewCache[string, int](ctx,
		OptionDefaultTTL(time.Hour),
		OptionMaxEntries(100),
		OptionTouchOnHit(false),
		OptionReaperInterval(10*time.Minute),
		OptionAllowEviction(false),
		OptionEvictionFn(func(key interface{}, entry interface{}) {}),
		OptionReaperBufferSize(200),
	)
	assert.Equal(t, time.Hour, cache.config.defaultTTL)
	assert.Equal(t, 100, cache.config.maxEntries)
	assert.Equal(t, false, cache.config.touchOnHit)
	assert.Equal(t, 10*time.Minute, cache.config.reaperInterval)
	assert.Equal(t, false, cache.config.allowEviction)
	assert.NotNil(t, cache.config.evictionFn)
	assert.Equal(t, 200, cache.config.reaperBufferSize)
	cache.Close(false)
}

func TestNewCacheOptionMaxEntriesRange(t *testing.T) {
	ctx := context.Background()
	t.Run("too low", func(tt *testing.T) {
		cache := NewCache[string, int](ctx,
			OptionMaxEntries(0),
		)
		assert.Equal(t, 1, cache.config.maxEntries)
		cache.Close(false)
	})
	t.Run("too high", func(tt *testing.T) {
		cache := NewCache[string, int](ctx,
			OptionMaxEntries(1001),
		)
		assert.Equal(t, 1000, cache.config.maxEntries)
		cache.Close(false)
	})
}

func TestNewCacheOptionReaperIntervalRange(t *testing.T) {
	ctx := context.Background()
	t.Run("too low", func(tt *testing.T) {
		cache := NewCache[string, int](ctx,
			OptionReaperInterval(time.Second),
		)
		assert.Equal(t, 30*time.Second, cache.config.reaperInterval)
		cache.Close(false)
	})
	t.Run("too high", func(tt *testing.T) {
		cache := NewCache[string, int](ctx,
			OptionReaperInterval(7201*time.Second),
		)
		assert.Equal(t, time.Hour, cache.config.reaperInterval)
		cache.Close(false)
	})
}

func TestNewCacheOptionReaperBufferSizeRange(t *testing.T) {
	ctx := context.Background()
	t.Run("too low", func(tt *testing.T) {
		cache := NewCache[string, int](ctx,
			OptionReaperBufferSize(-1),
		)
		assert.Equal(t, 1, cache.config.reaperBufferSize)
		cache.Close(false)
	})
}

func TestNewCacheOptionDefaultTTLRange(t *testing.T) {
	ctx := context.Background()
	t.Run("too low", func(tt *testing.T) {
		cache := NewCache[string, int](ctx,
			OptionDefaultTTL(30*time.Second),
		)
		assert.Equal(t, time.Minute, cache.config.defaultTTL)
		cache.Close(false)
	})
	t.Run("too high", func(tt *testing.T) {
		cache := NewCache[string, int](ctx,
			OptionDefaultTTL(25*time.Hour),
		)
		assert.Equal(t, 24*time.Hour, cache.config.defaultTTL)
		cache.Close(false)
	})
}
