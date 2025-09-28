package champ

import (
	"fmt"
	"iter"
	"slices"
)

type node[K comparable, V any] interface {
	fmt.Stringer

	get(key K, hash uint64, shift uint) (V, bool)
	set(key K, value V, hash uint64, shift uint, hashFunc func(key K) uint64) (node[K, V], bool)
	del(key K, hash uint64, shift uint) (node[K, V], bool)
	all() iter.Seq2[K, V]
	keysSeq() iter.Seq[K]
	valuesSeq() iter.Seq[V]
}

// bitmapIndexedNode is the main CHAMP node type with compressed storage
type bitmapIndexedNode[K comparable, V any] struct {
	nodemap uint32       // Bitmap for child nodes
	datamap uint32       // Bitmap for key-value pairs
	nodes   []node[K, V] // Array of child nodes (compressed)
	keys    []K          // Array of keys (compressed)
	values  []V          // Array of values (compressed)
}

func (n *bitmapIndexedNode[K, V]) get(key K, hash uint64, shift uint) (V, bool) {
	var zero V
	bit := uint32(1 << ((hash >> shift) & bitMask))

	if n.datamap&bit != 0 {
		idx := popcount(n.datamap & (bit - 1))
		if n.keys[idx] == key {
			return n.values[idx], true
		}
		return zero, false
	}

	if n.nodemap&bit != 0 {
		idx := popcount(n.nodemap & (bit - 1))
		return n.nodes[idx].get(key, hash, shift+bitsPerLevel)
	}

	return zero, false
}

func (n *bitmapIndexedNode[K, V]) set(key K, value V, hash uint64, shift uint, hashFunc func(key K) uint64) (node[K, V], bool) {
	bit := uint32(1 << ((hash >> shift) & bitMask))

	if n.datamap&bit != 0 {
		idx := popcount(n.datamap & (bit - 1))
		if n.keys[idx] == key {
			newKeys := make([]K, len(n.keys))
			copy(newKeys, n.keys)
			newValues := make([]V, len(n.values))
			copy(newValues, n.values)
			newValues[idx] = value

			return &bitmapIndexedNode[K, V]{
				nodemap: n.nodemap,
				datamap: n.datamap,
				nodes:   n.nodes,
				keys:    newKeys,
				values:  newValues,
			}, false
		}

		// Collision
		subNode := n.createSubNode(
			n.keys[idx], n.values[idx], hashFunc(n.keys[idx]),
			key, value, hash,
			shift+bitsPerLevel,
		)
		return &bitmapIndexedNode[K, V]{
			nodemap: n.nodemap | bit,
			datamap: n.datamap &^ bit,
			nodes:   insertAt(n.nodes, popcount(n.nodemap&(bit-1)), subNode),
			keys:    removeAt(n.keys, idx),
			values:  removeAt(n.values, idx),
		}, true
	}

	// Check if we have a node at this position
	if n.nodemap&bit != 0 {
		idx := popcount(n.nodemap & (bit - 1))
		newNode, added := n.nodes[idx].set(key, value, hash, shift+bitsPerLevel, hashFunc)
		if !added && newNode == n.nodes[idx] {
			return n, false
		}

		newNodes := make([]node[K, V], len(n.nodes))
		copy(newNodes, n.nodes)
		newNodes[idx] = newNode

		return &bitmapIndexedNode[K, V]{
			nodemap: n.nodemap,
			datamap: n.datamap,
			nodes:   newNodes,
			keys:    n.keys,
			values:  n.values,
		}, added
	}

	// Empty position
	idx := popcount(n.datamap & (bit - 1))
	return &bitmapIndexedNode[K, V]{
		nodemap: n.nodemap,
		datamap: n.datamap | bit,
		nodes:   n.nodes,
		keys:    insertAt(n.keys, idx, key),
		values:  insertAt(n.values, idx, value),
	}, true
}

func (n *bitmapIndexedNode[K, V]) del(key K, hash uint64, shift uint) (node[K, V], bool) {
	bit := uint32(1 << ((hash >> shift) & bitMask))

	if n.datamap&bit != 0 {
		idx := popcount(n.datamap & (bit - 1))
		if n.keys[idx] != key {
			return n, false
		}

		if len(n.keys) == 1 && len(n.nodes) == 0 {
			return nil, true
		}

		return &bitmapIndexedNode[K, V]{
			nodemap: n.nodemap,
			datamap: n.datamap &^ bit,
			nodes:   n.nodes,
			keys:    removeAt(n.keys, idx),
			values:  removeAt(n.values, idx),
		}, true
	}

	if n.nodemap&bit != 0 {
		idx := popcount(n.nodemap & (bit - 1))
		newNode, deleted := n.nodes[idx].del(key, hash, shift+bitsPerLevel)

		if !deleted {
			return n, false
		}

		if newNode == nil {
			// Remove empty node
			if len(n.nodes) == 1 && len(n.keys) == 0 {
				return nil, true
			}

			return &bitmapIndexedNode[K, V]{
				nodemap: n.nodemap &^ bit,
				datamap: n.datamap,
				nodes:   removeAt(n.nodes, idx),
				keys:    n.keys,
				values:  n.values,
			}, true
		}

		if m, ok := newNode.(*bitmapIndexedNode[K, V]); ok && m.nodemap == 0 && len(m.keys) == 1 {
			// Collapse single entry node
			return &bitmapIndexedNode[K, V]{
				nodemap: n.nodemap &^ bit,
				datamap: n.datamap | bit,
				nodes:   removeAt(n.nodes, idx),
				keys:    insertAt(n.keys, popcount(n.datamap&(bit-1)), m.keys[0]),
				values:  insertAt(n.values, popcount(n.datamap&(bit-1)), m.values[0]),
			}, true
		}

		newNodes := make([]node[K, V], len(n.nodes))
		copy(newNodes, n.nodes)
		newNodes[idx] = newNode

		return &bitmapIndexedNode[K, V]{
			nodemap: n.nodemap,
			datamap: n.datamap,
			nodes:   newNodes,
			keys:    n.keys,
			values:  n.values,
		}, true
	}

	return n, false
}

func (n *bitmapIndexedNode[K, V]) all() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for i := 0; i < len(n.keys); i++ {
			if !yield(n.keys[i], n.values[i]) {
				return
			}
		}
		for _, child := range n.nodes {
			for k, v := range child.all() {
				if !yield(k, v) {
					return
				}
			}
		}
	}
}

func (n *bitmapIndexedNode[K, V]) keysSeq() iter.Seq[K] {
	return func(yield func(K) bool) {
		for _, k := range n.keys {
			if !yield(k) {
				return
			}
		}
		for _, child := range n.nodes {
			for k := range child.keysSeq() {
				if !yield(k) {
					return
				}
			}
		}
	}
}

func (n *bitmapIndexedNode[K, V]) valuesSeq() iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, v := range n.values {
			if !yield(v) {
				return
			}
		}
		for _, child := range n.nodes {
			for v := range child.valuesSeq() {
				if !yield(v) {
					return
				}
			}
		}
	}
}

func (n *bitmapIndexedNode[K, V]) String() string {
	return bitmapIndexedNodeString(n, 0, 0)
}

func (n *bitmapIndexedNode[K, V]) createSubNode(
	key1 K, val1 V, hash1 uint64,
	key2 K, val2 V, hash2 uint64,
	shift uint,
) node[K, V] {
	if shift >= maxDepth*bitsPerLevel {
		// Create collision node
		return &collisionNode[K, V]{
			keys:   []K{key1, key2},
			values: []V{val1, val2},
		}
	}

	bit1 := uint32(1 << ((hash1 >> shift) & bitMask))
	bit2 := uint32(1 << ((hash2 >> shift) & bitMask))

	if bit1 == bit2 {
		// Same position at this level, recurse
		subNode := n.createSubNode(key1, val1, hash1, key2, val2, hash2, shift+bitsPerLevel)
		return &bitmapIndexedNode[K, V]{
			nodemap: bit1,
			nodes:   []node[K, V]{subNode},
		}
	}

	if bit1 < bit2 {
		return &bitmapIndexedNode[K, V]{
			datamap: bit1 | bit2,
			keys:    []K{key1, key2},
			values:  []V{val1, val2},
		}
	}
	return &bitmapIndexedNode[K, V]{
		datamap: bit1 | bit2,
		keys:    []K{key2, key1},
		values:  []V{val2, val1},
	}
}

// collisionNode handles hash collisions
type collisionNode[K comparable, V any] struct {
	keys   []K
	values []V
}

func (n *collisionNode[K, V]) get(key K, hash uint64, shift uint) (V, bool) {
	for i, k := range n.keys {
		if k == key {
			return n.values[i], true
		}
	}
	var zero V
	return zero, false
}

func (n *collisionNode[K, V]) set(key K, value V, hash uint64, shift uint, _ func(key K) uint64) (node[K, V], bool) {
	for i, k := range n.keys {
		if k == key {
			// Update existing
			newValues := make([]V, len(n.values))
			copy(newValues, n.values)
			newValues[i] = value

			return &collisionNode[K, V]{
				keys:   n.keys,
				values: newValues,
			}, false
		}
	}

	// Add new entry at the end
	newKeys := make([]K, len(n.keys)+1)
	newValues := make([]V, len(n.values)+1)
	copy(newKeys, n.keys)
	copy(newValues, n.values)
	newKeys[len(n.keys)] = key
	newValues[len(n.values)] = value

	return &collisionNode[K, V]{
		keys:   newKeys,
		values: newValues,
	}, true
}

func (n *collisionNode[K, V]) del(key K, hash uint64, shift uint) (node[K, V], bool) {
	for i, k := range n.keys {
		if k == key {
			if len(n.keys) == 2 {
				// Convert to bitmapIndexedNode when only one entry remains.
				// The parent bitmapIndexedNode will detect this single-entry node
				// and collapse it into its own data array.
				// We use an arbitrary bit position (0) since this node will be collapsed anyway.
				return &bitmapIndexedNode[K, V]{
					datamap: 1, // Set first bit to indicate one data entry
					keys:    []K{n.keys[1-i]},
					values:  []V{n.values[1-i]},
				}, true
			}

			return &collisionNode[K, V]{
				keys:   removeAt(n.keys, i),
				values: removeAt(n.values, i),
			}, true
		}
	}
	return n, false
}

func (n *collisionNode[K, V]) all() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for i := 0; i < len(n.keys); i++ {
			if !yield(n.keys[i], n.values[i]) {
				return
			}
		}
	}
}

func (n *collisionNode[K, V]) keysSeq() iter.Seq[K] {
	return slices.Values(n.keys)
}

func (n *collisionNode[K, V]) valuesSeq() iter.Seq[V] {
	return slices.Values(n.values)
}

func (n *collisionNode[K, V]) String() string {
	return collisionNodeString(n, 0)
}

// popcount returns the number of set bits in x.
func popcount(x uint32) int {
	x = x - ((x >> 1) & 0x55555555)
	x = (x & 0x33333333) + ((x >> 2) & 0x33333333)
	x = (x + (x >> 4)) & 0x0f0f0f0f
	x = x + (x >> 8)
	x = x + (x >> 16)
	return int(x & 0x3f)
}

func insertAt[T any](arr []T, idx int, val T) []T {
	result := make([]T, len(arr)+1)
	copy(result[:idx], arr[:idx])
	result[idx] = val
	copy(result[idx+1:], arr[idx:])
	return result
}

func removeAt[T any](arr []T, idx int) []T {
	result := make([]T, len(arr)-1)
	copy(result[:idx], arr[:idx])
	copy(result[idx:], arr[idx+1:])
	return result
}
