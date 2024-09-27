/*
 * Copyright (C) 2024 by Jason Figge
 */

package cache

import (
	"sync"
)

type queue[T comparable] struct {
	lock       sync.Mutex
	items      []T
	count      int
	head       int
	tail       int
	expandable bool
	null       T
}

func newQueue[T comparable]() *queue[T] {
	return &queue[T]{
		items:      make([]T, 100),
		expandable: true,
	}
}

func newQueueWithCapacity[T comparable](capacity int) *queue[T] {
	return &queue[T]{
		items:      make([]T, capacity),
		expandable: false,
	}
}

func (q *queue[T]) Len() int {
	return q.count
}

func (q *queue[T]) Cap() int {
	return cap(q.items)
}

func (q *queue[T]) Remove(item T) bool {
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
	for i := 0; i < q.count; i++ {
		if q.items[i] == item {
			q.count--
			q.head--
			copy(q.items[i:], q.items[i+1:])
			q.items[q.count] = q.null
			return true
		}
	}
	return false
}

func (q *queue[T]) Push(item T) bool {
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

func (q *queue[T]) Pop() (T, bool) {
	q.lock.Lock()
	defer q.lock.Unlock()
	if q.tail != q.head || q.count > 0 {
		item := q.items[q.tail]
		q.items[q.tail] = q.null
		q.tail = q.next(q.tail)
		q.count--
		return item, true
	}
	return q.null, false
}

func (q *queue[T]) next(p int) int {
	p++
	if p == cap(q.items) {
		return 0
	}
	return p
}

func (q *queue[T]) expandCapacity() {
	temp := make([]T, q.count+100)
	copy(temp, q.items[q.tail:])
	copy(temp[q.count-q.head:q.count], q.items[:q.head])
	q.items = temp
	q.tail = 0
	q.head = q.count
}

func (q *queue[T]) Items() []T {
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
