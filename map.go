package champ

import (
	"hash/fnv"
	"iter"
)

const (
	bitsPerLevel = 5
	branchFactor = 1 << bitsPerLevel
	bitMask      = branchFactor - 1
	maxDepth     = 7 // 32 bits / 5 bits per level = 6.4
)

type Key interface {
	~string
}

// Map represents a CHAMP (Compressed Hash-Array Mapped Prefix-tree)
type Map[K Key, V any] struct {
	root node[K, V]
	size int
}

// New creates a new empty CHAMP map.
func New[K Key, V any]() *Map[K, V] {
	return &Map[K, V]{
		root: nil,
		size: 0,
	}
}

// Get retrieves a value by key.
func (m *Map[K, V]) Get(key K) (V, bool) {
	var zero V
	if m.root == nil {
		return zero, false
	}
	return m.root.get(key, hashKey(key), 0)
}

// Set sets or updates a key-value pair.
func (m *Map[K, V]) Set(key K, value V) *Map[K, V] {
	h := hashKey(key)

	if m.root == nil {
		return &Map[K, V]{
			root: &bitmapIndexedNode[K, V]{
				datamap: uint32(1 << (h & bitMask)),
				keys:    []K{key},
				values:  []V{value},
			},
			size: 1,
		}
	}

	root, added := m.root.set(key, value, h, 0, hashKey)
	size := m.size
	if added {
		size++
	}

	return &Map[K, V]{
		root: root,
		size: size,
	}
}

// Delete removes a key from the map.
func (m *Map[K, V]) Delete(key K) *Map[K, V] {
	if m.root == nil {
		return m
	}

	newRoot, deleted := m.root.del(key, hashKey(key), 0)
	if !deleted {
		return m
	}

	newSize := m.size - 1
	if newRoot == nil {
		return New[K, V]()
	}

	return &Map[K, V]{
		root: newRoot,
		size: newSize,
	}
}

// Len returns the number of entries
func (m *Map[K, V]) Len() int {
	return m.size
}

// All returns an iterator over key-value pairs.
func (m *Map[K, V]) All() iter.Seq2[K, V] {
	if m.root != nil {
		return m.root.all()
	}
	return func(func(K, V) bool) {}
}

func hashKey[K Key](key K) uint32 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32()
}

func Equal[K Key, V comparable](m1, m2 *Map[K, V]) bool {
	if m1.size != m2.size {
		return false
	}
	return equalNode(m1.root, m2.root)
}

func equalNode[K Key, V comparable](n1, n2 node[K, V]) bool {
	// short-circuit for identical pointers
	if n1 == n2 {
		return true
	}

	switch n1 := n1.(type) {
	case *bitmapIndexedNode[K, V]:
		n2, ok := n2.(*bitmapIndexedNode[K, V])
		if !ok {
			return false
		}
		return equalBitmapIndexedNodes(n1, n2)
	case *collisionNode[K, V]:
		n2, ok := n2.(*collisionNode[K, V])
		if !ok {
			return false
		}
		return equalCollisionNodes(n1, n2)
	}

	return false
}

func equalBitmapIndexedNodes[K Key, V comparable](n1, n2 *bitmapIndexedNode[K, V]) bool {
	if n1.datamap != n2.datamap || n1.nodemap != n2.nodemap {
		return false
	}
	for i := range n1.keys {
		if n1.keys[i] != n2.keys[i] || n1.values[i] != n2.values[i] {
			return false
		}
	}
	for i := range n1.nodes {
		if !equalNode(n1.nodes[i], n2.nodes[i]) {
			return false
		}
	}
	return true
}

func equalCollisionNodes[K Key, V comparable](n1, n2 *collisionNode[K, V]) bool {
	if len(n1.keys) != len(n2.keys) {
		return false
	}
	for i := range n1.keys {
		if n1.keys[i] != n2.keys[i] || n1.values[i] != n2.values[i] {
			return false
		}
	}
	return true
}
