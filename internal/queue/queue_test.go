/*
 * Copyright (C) 2024 by Jason Figge
 */

package queue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigDefault(t *testing.T) {
	q := NewQueue[int]()
	assert.Equal(t, 100, q.config.capacity)
	assert.False(t, q.config.compactOnRemove)
	assert.True(t, q.config.expandable)
}

func TestConfigOptions(t *testing.T) {
	q := NewQueue[int](
		OptionCapacity(20),
		OptionCompactOnRemove(true),
	)
	assert.Equal(t, 20, q.config.capacity)
	assert.False(t, q.config.expandable)
	assert.True(t, q.config.compactOnRemove)
}

func TestConfigOptionsCapacityRange(t *testing.T) {
	q := NewQueue[int](OptionCapacity(4))
	assert.Equal(t, 10, q.config.capacity)
}

func TestFixedCapacity(t *testing.T) {
	minCapacity = 4
	q := NewQueue[int](OptionCapacity(4))
	assert.Equal(t, 4, q.Cap())
	assert.True(t, q.Push(1))
	assert.Equal(t, 1, q.Len())
	assert.True(t, q.Push(2))
	assert.Equal(t, 2, q.Len())
	assert.True(t, q.Push(3))
	assert.Equal(t, 3, q.Len())
	assert.True(t, q.Push(4))
	assert.Equal(t, 4, q.Len())
	assert.False(t, q.Push(5))
	assert.Equal(t, 4, q.Len())

	item, ok := q.Pop()
	assert.True(t, ok)
	assert.Equal(t, 1, item)
	assert.Equal(t, 3, q.Len())
	//assert.ElementsMatch(t, []int{2, 3, 4}, q.Items())
	item, ok = q.Pop()
	assert.True(t, ok)
	assert.Equal(t, 2, item)
	assert.Equal(t, 2, q.Len())
	item, ok = q.Pop()
	assert.True(t, ok)
	assert.Equal(t, 3, item)
	assert.Equal(t, 1, q.Len())
	item, ok = q.Pop()
	assert.True(t, ok)
	assert.Equal(t, 4, item)
	assert.Equal(t, 0, q.Len())
	item, ok = q.Pop()
	assert.False(t, ok)
	assert.Equal(t, 0, item)
	assert.Equal(t, 0, q.Len())

	q.Push(5)
	q.Push(6)
	q.Push(7)
	assert.Equal(t, 3, q.Len())
	q.Pop()
	q.Pop()
	assert.Equal(t, 1, q.Len())
	q.Push(8)
	q.Push(9)
	assert.Equal(t, 3, q.Len())
	q.Pop()
	q.Pop()
	item, ok = q.Pop()
	assert.True(t, ok)
	assert.Equal(t, 9, item)
	assert.Equal(t, 0, q.Len())
	item, ok = q.Pop()
	assert.False(t, ok)
	assert.Equal(t, 0, item)
	assert.Equal(t, 0, q.Len())
}

func TestDynamicCapacity(t *testing.T) {
	q := NewQueue[int]()
	assert.Equal(t, 100, q.Cap())
	for i := 0; i < 1000; i++ {
		assert.True(t, q.Push(i+1))
		assert.Equal(t, i+1, q.Len())
	}
	for i := 0; q.Len() > 0; i++ {
		item, ok := q.Pop()
		assert.True(t, ok)
		assert.Equal(t, i+1, item)
	}
	item, ok := q.Pop()
	assert.False(t, ok)
	assert.Equal(t, 0, item)
	assert.Equal(t, 0, q.Len())
}

func TestDynamic2Capacity(t *testing.T) {
	q := NewQueue[int]()
	assert.Equal(t, 100, q.Cap())
	in := 1
	for i := 0; i < 70; i++ {
		q.Push(in)
		in++
	}
	out := 1
	for i := 0; q.Len() > 40; i++ {
		item, ok := q.Pop()
		assert.True(t, ok)
		assert.Equal(t, out, item)
		out++
	}
	for i := 0; i < 80; i++ {
		q.Push(in)
		in++
	}
	for i := 0; q.Len() > 0; i++ {
		item, ok := q.Pop()
		assert.True(t, ok)
		assert.Equal(t, out, item)
		out++
	}
	item, ok := q.Pop()
	assert.False(t, ok)
	assert.Equal(t, 0, item)
	assert.Equal(t, 0, q.Len())
}

func TestItems(t *testing.T) {
	q := NewQueue[int](OptionCapacity(4))
	q.Push(1)
	assert.ElementsMatch(t, []int{1}, q.Items())
	q.Push(2)
	assert.ElementsMatch(t, []int{1, 2}, q.Items())
	q.Push(3)
	assert.ElementsMatch(t, []int{1, 2, 3}, q.Items())
	q.Pop()
	assert.ElementsMatch(t, []int{2, 3}, q.Items())
	q.Pop()
	assert.ElementsMatch(t, []int{3}, q.Items())
	q.Push(4)
	assert.ElementsMatch(t, []int{3, 4}, q.Items())
	q.Push(5)
	assert.ElementsMatch(t, []int{3, 4, 5}, q.Items())
	q.Push(6)
	assert.ElementsMatch(t, []int{3, 4, 5, 6}, q.Items())
	q.Pop()
	assert.ElementsMatch(t, []int{4, 5, 6}, q.Items())
	q.Pop()
	assert.ElementsMatch(t, []int{5, 6}, q.Items())
	q.Pop()
	assert.ElementsMatch(t, []int{6}, q.Items())
	q.Pop()
	assert.ElementsMatch(t, []int{}, q.Items())
}

func TestRemoveCompact(t *testing.T) {
	minCapacity = 4
	q := NewQueue[int](OptionCapacity(4), OptionCompactOnRemove(true))
	q.Push(1)
	q.Push(2)
	q.Push(3)
	q.Push(4)
	assert.ElementsMatch(t, []int{1, 2, 3, 4}, q.Items())
	q.Remove(2)
	assert.ElementsMatch(t, []int{1, 3, 4}, q.Items())
	q.Remove(1)
	assert.ElementsMatch(t, []int{3, 4}, q.Items())
	q.Remove(4)
	assert.ElementsMatch(t, []int{3}, q.Items())
	q.Remove(3)
	assert.ElementsMatch(t, []int{}, q.Items())
	q.Push(5)
	q.Push(6)
	q.Push(7)
	q.Push(8)
	q.Pop()
	q.Pop()
	assert.ElementsMatch(t, []int{7, 8}, q.Items())
	q.Push(9)
	assert.ElementsMatch(t, []int{7, 8, 9}, q.Items())
	q.Pop()
	assert.ElementsMatch(t, []int{8, 9}, q.Items())
	q.Push(10)
	q.Push(11)
	assert.ElementsMatch(t, []int{8, 9, 10, 11}, q.Items())
	q.Remove(10)
	assert.ElementsMatch(t, []int{8, 9, 11}, q.Items())
	assert.Equal(t, 1, q.Remove(8))
	assert.ElementsMatch(t, []int{9, 11}, q.Items())
	assert.Equal(t, 0, q.Remove(8))
	assert.ElementsMatch(t, []int{9, 11}, q.Items())
	q.Remove(11)
	assert.ElementsMatch(t, []int{9}, q.Items())
	q.Remove(9)
	assert.ElementsMatch(t, []int{}, q.Items())
	q.Push(12)
	q.Push(13)
	q.Push(14)
	q.Pop()
	assert.ElementsMatch(t, []int{13, 14}, q.Items())
	q.Remove(13)
	assert.ElementsMatch(t, []int{14}, q.Items())
	q.Remove(14)
	assert.ElementsMatch(t, []int{}, q.Items())
	assert.Equal(t, 0, q.removed)
}

func TestRemoveNoCompact(t *testing.T) {
	minCapacity = 4
	q := NewQueue[int](OptionCapacity(4))
	q.Push(1)
	q.Push(2)
	q.Push(3)
	q.Push(4)
	assert.ElementsMatch(t, []int{1, 2, 3, 4}, q.Items())
	q.Remove(2)
	assert.ElementsMatch(t, []int{1, 0, 3, 4}, q.Items())
	q.Remove(1)
	assert.ElementsMatch(t, []int{0, 0, 3, 4}, q.Items())
	q.Remove(4)
	assert.ElementsMatch(t, []int{0, 0, 3, 0}, q.Items())
	q.Remove(3)
	assert.ElementsMatch(t, []int{}, q.Items())
	q.Push(5)
	q.Push(6)
	q.Push(7)
	q.Push(8)
	q.Pop()
	q.Pop()
	assert.ElementsMatch(t, []int{7, 8}, q.Items())
	q.Push(9)
	assert.ElementsMatch(t, []int{7, 8, 9}, q.Items())
	q.Pop()
	assert.ElementsMatch(t, []int{8, 9}, q.Items())
	q.Push(10)
	q.Push(11)
	assert.ElementsMatch(t, []int{8, 9, 10, 11}, q.Items())
	q.Remove(10)
	assert.ElementsMatch(t, []int{8, 9, 0, 11}, q.Items())
	assert.Equal(t, 1, q.Remove(8))
	assert.ElementsMatch(t, []int{0, 9, 0, 11}, q.Items())
	assert.Equal(t, 0, q.Remove(8))
	assert.ElementsMatch(t, []int{0, 9, 0, 11}, q.Items())
	q.Remove(11)
	assert.ElementsMatch(t, []int{0, 9, 0, 0}, q.Items())
	q.Remove(9)
	assert.ElementsMatch(t, []int{}, q.Items())
	q.Push(12)
	q.Push(13)
	q.Push(14)
	q.Pop()
	assert.ElementsMatch(t, []int{13, 14}, q.Items())
	q.Remove(13)
	assert.ElementsMatch(t, []int{0, 14}, q.Items())
	q.Remove(14)
	assert.ElementsMatch(t, []int{}, q.Items())
	assert.Equal(t, 0, q.removed)
}
