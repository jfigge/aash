/*
 * Copyright (C) 2024 by Jason Figge
 */

package queue

import (
	"sync"
)

var (
	minCapacity = 10
)

type OptFn func(*config)

type config struct {
	capacity        int
	expandable      bool
	compactOnRemove bool
}

type Queue[T comparable] struct {
	lock    sync.Mutex
	items   []T
	count   int
	head    int
	tail    int
	null    T
	removed int
	*config
}

func NewQueue[T comparable](options ...OptFn) *Queue[T] {
	q := &Queue[T]{
		config: &config{
			expandable:      true,
			capacity:        100,
			compactOnRemove: false,
		},
	}
	for _, option := range options {
		option(q.config)
	}
	q.items = make([]T, q.capacity)

	return q
}

func OptionCapacity(capacity int) OptFn {
	return func(c *config) {
		if capacity < minCapacity {
			capacity = minCapacity
		}
		c.expandable = false
		c.capacity = capacity
	}
}
func OptionCompactOnRemove(compactOnRemove bool) OptFn {
	return func(c *config) {
		c.compactOnRemove = compactOnRemove
	}
}

func (q *Queue[T]) Len() int {
	return q.count
}

func (q *Queue[T]) Cap() int {
	return cap(q.items)
}

func (q *Queue[T]) Remove(item T) int {
	removed := 0
	for i := 0; i < len(q.items); i++ {
		if q.items[i] == item {
			q.items[i] = q.null
			removed++
			q.removed++
		}
	}
	if q.compactOnRemove {
		q.Compact()
	} else if q.count == q.removed {
		q.head = 0
		q.tail = 0
		q.count = 0
		q.removed = 0
	}
	return removed
}

func (q *Queue[T]) Compact() {
	q.lock.Lock()
	defer q.lock.Unlock()
	if q.tail >= q.head && q.tail != 0 {
		temp := make([]T, cap(q.items))
		copy(temp, q.items[q.tail:])
		copy(temp[cap(q.items)-q.tail:], q.items[:q.head])
		q.items = temp
	} else if q.tail > 0 {
		temp := make([]T, q.count)
		copy(temp, q.items[q.tail:q.head])
		q.items = temp
	}
	q.tail = 0
	q.head = q.count
	for i := 0; i < len(q.items) && q.removed > 0; i++ {
		if q.items[i] == q.null {
			q.count--
			q.head--
			q.removed--
			copy(q.items[i:], q.items[i+1:])
			q.items[q.count] = q.null
		}
	}
}

func (q *Queue[T]) Push(item T) bool {
	q.lock.Lock()
	defer q.lock.Unlock()
	if cap(q.items) == q.count {
		if q.expandable {
			q.expandCapacity()
		} else {
			return false
		}
	}
	q.items[q.head] = item
	q.head = q.next(q.head)
	q.count++
	return true
}

func (q *Queue[T]) Pop() (T, bool) {
	q.lock.Lock()
	defer q.lock.Unlock()
	for q.tail != q.head || q.count > 0 {
		item := q.items[q.tail]
		q.items[q.tail] = q.null
		q.tail = q.next(q.tail)
		q.count--
		if item == q.null && q.removed > 0 {
			q.removed--
			continue
		}
		return item, true
	}
	return q.null, false
}

func (q *Queue[T]) next(p int) int {
	p++
	if p == cap(q.items) {
		return 0
	}
	return p
}

func (q *Queue[T]) expandCapacity() {
	temp := make([]T, q.count+100)
	copy(temp, q.items[q.tail:])
	copy(temp[q.count-q.head:q.count], q.items[:q.head])
	q.items = temp
	q.tail = 0
	q.head = q.count
}

func (q *Queue[T]) Items() []T {
	q.lock.Lock()
	defer q.lock.Unlock()
	temp := make([]T, q.count)
	if q.count == 0 {
		return temp
	}
	if q.tail < q.head {
		copy(temp, q.items[q.tail:q.head])
	} else {
		copy(temp, q.items[q.tail:])
		copy(temp[q.count-q.head:q.count], q.items[:q.head])
	}
	return temp
}
