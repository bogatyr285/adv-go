package hw4_test

import (
	"fmt"
	"reflect"
	"testing"

	"cmp"

	"github.com/stretchr/testify/assert"
)

type OrderedMap[K cmp.Ordered, V any] struct {
	root *node[K, V]
	size int
}

type node[K cmp.Ordered, V any] struct {
	left, right *node[K, V]
	key         K
	value       V
}

func New[K cmp.Ordered, V any]() *OrderedMap[K, V] {
	return &OrderedMap[K, V]{root: nil, size: 0}
}

func (m *OrderedMap[K, V]) Insert(key K, value V) {
	if m.root == nil {
		m.root = &node[K, V]{key: key, value: value}
		m.size++
		return
	}

	current := m.root
	for {
		switch {
		case key == current.key:
			current.value = value
			return
		case cmp.Less(key, current.key):
			if current.left == nil {
				current.left = &node[K, V]{key: key, value: value}
				m.size++
				return
			}
			current = current.left
		default:
			if current.right == nil {
				current.right = &node[K, V]{key: key, value: value}
				m.size++
				return
			}
			current = current.right
		}
	}
}

func (m *OrderedMap[K, V]) Erase(key K) {
	m.root = m.erase(m.root, key)
}

func (m *OrderedMap[K, V]) erase(n *node[K, V], key K) *node[K, V] {
	if n == nil {
		return nil
	}

	switch {
	case cmp.Less(key, n.key):
		n.left = m.erase(n.left, key)
	case cmp.Less(n.key, key):
		n.right = m.erase(n.right, key)
	default:
		if n.left == nil {
			m.size--
			return n.right
		}
		if n.right == nil {
			m.size--
			return n.left
		}

		minRight := m.minnode(n.right)
		n.key, n.value = minRight.key, minRight.value
		n.right = m.erase(n.right, minRight.key)
	}
	return n
}

func (m *OrderedMap[K, V]) Contains(key K) bool {
	current := m.root
	for current != nil {
		switch {
		case key == current.key:
			return true
		case cmp.Less(key, current.key):
			current = current.left
		default:
			current = current.right
		}
	}
	return false
}

func (m *OrderedMap[K, V]) Size() int {
	return m.size
}

func (m *OrderedMap[K, V]) ForEach(f func(K, V)) {
	m.traverse(m.root, f)
}

func (m *OrderedMap[K, V]) minnode(n *node[K, V]) *node[K, V] {
	current := n
	for current != nil && current.left != nil {
		current = current.left
	}
	return current
}

func (m *OrderedMap[K, V]) traverse(n *node[K, V], f func(K, V)) {
	if n == nil {
		return
	}
	m.traverse(n.left, f)
	f(n.key, n.value)
	m.traverse(n.right, f)
}

func TestCircularQueue(t *testing.T) {
	data := New[int, int]()
	assert.Zero(t, data.Size())

	data.Insert(10, 10)
	data.Insert(5, 5)
	data.Insert(15, 15)
	data.Insert(2, 2)
	data.Insert(4, 4)
	data.Insert(12, 12)
	data.Insert(14, 14)

	assert.Equal(t, 7, data.Size())
	assert.True(t, data.Contains(4))
	assert.True(t, data.Contains(12))
	assert.False(t, data.Contains(3))
	assert.False(t, data.Contains(13))

	var keys []int
	expectedKeys := []int{2, 4, 5, 10, 12, 14, 15}
	data.ForEach(func(key, _ int) {
		keys = append(keys, key)
	})

	assert.True(t, reflect.DeepEqual(expectedKeys, keys))

	data.Erase(15)
	data.Erase(14)
	data.Erase(2)

	assert.Equal(t, 4, data.Size())
	assert.True(t, data.Contains(4))
	assert.True(t, data.Contains(12))
	assert.False(t, data.Contains(2))
	assert.False(t, data.Contains(14))

	keys = nil
	expectedKeys = []int{4, 5, 10, 12}
	data.ForEach(func(key, _ int) {
		keys = append(keys, key)
	})

	assert.True(t, reflect.DeepEqual(expectedKeys, keys))
}

func BenchmarkOrderedMap(b *testing.B) {
	b.Run("Insert", func(b *testing.B) {
		m := New[int, string]()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			m.Insert(i, fmt.Sprintf("value%d", i))
		}
	})

	b.Run("Contains", func(b *testing.B) {
		m := New[int, string]()
		for i := 0; i < 1000; i++ {
			m.Insert(i, fmt.Sprintf("value%d", i))
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			m.Contains(i % 1000)
		}
	})

	b.Run("Erase", func(b *testing.B) {
		m := New[int, string]()
		for i := 0; i < 1000; i++ {
			m.Insert(i, fmt.Sprintf("value%d", i))
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			m.Erase(i % 1000)
			b.StopTimer()
			// вставляем столько же сколько удалили
			m.Insert(i%1000, fmt.Sprintf("value%d", i%1000))
			b.StartTimer()
		}
	})

	b.Run("ForEach", func(b *testing.B) {
		m := New[int, string]()
		for i := 0; i < 1000; i++ {
			m.Insert(i, fmt.Sprintf("value%d", i))
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			m.ForEach(func(k int, v string) {})
		}
	})
}
