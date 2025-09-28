package champ

import (
	"strconv"
	"strings"
	"testing"
)

// Simple hash function for testing, which converts binary string to uint64.
func testHashFunc(key string) uint64 {
	i, err := strconv.ParseUint(key, 2, 64)
	if err != nil {
		panic(err)
	}
	return i
}

func TestBitmapIndexedNode(t *testing.T) {
	t.Run("get", func(t *testing.T) {
		for _, tt := range []struct {
			name       string
			node       *bitmapIndexedNode[string, int]
			key        string
			hash       uint64
			shift      uint
			expected   int
			expectedOk bool
		}{
			{
				name:       "empty node",
				node:       &bitmapIndexedNode[string, int]{},
				key:        "00001",
				hash:       0b00001,
				shift:      0,
				expected:   0,
				expectedOk: false,
			},
			{
				name: "single key present",
				node: &bitmapIndexedNode[string, int]{
					datamap: 0b00010, // bit 1 set
					keys:    []string{"00001"},
					values:  []int{100},
				},
				key:        "00001",
				hash:       0b00001,
				shift:      0,
				expected:   100,
				expectedOk: true,
			},
			{
				name: "multiple keys present",
				node: &bitmapIndexedNode[string, int]{
					datamap: 0b10101, // bits 0, 2, 4 set
					keys:    []string{"00000", "00010", "00100"},
					values:  []int{10, 30, 50},
				},
				key:        "00010",
				hash:       0b00010,
				shift:      0,
				expected:   30,
				expectedOk: true,
			},
			{
				name: "not found",
				node: &bitmapIndexedNode[string, int]{
					datamap: 0b10101, // bits 0, 2, 4 set
					keys:    []string{"00000", "00010", "00100"},
					values:  []int{10, 30, 50},
				},
				key:        "00011",
				hash:       0b00011,
				shift:      0,
				expected:   0,
				expectedOk: false,
			},
			{
				name: "child node present",
				node: &bitmapIndexedNode[string, int]{
					nodemap: 0b00001,
					nodes: []node[string, int]{
						&bitmapIndexedNode[string, int]{
							datamap: 0b00010,
							keys:    []string{"0000100000"},
							values:  []int{200},
						},
					},
				},
				key:        "0000100000",
				hash:       0b0000100000,
				shift:      0,
				expected:   200,
				expectedOk: true,
			},
		} {
			t.Run(tt.name, func(t *testing.T) {
				got, ok := tt.node.get(tt.key, tt.hash, tt.shift)
				if ok != tt.expectedOk {
					t.Errorf("get() ok = %v, expected %v", ok, tt.expectedOk)
				}
				if got != tt.expected {
					t.Errorf("get() value = %d, expected %d", got, tt.expected)
				}
			})
		}
	})

	t.Run("set", func(t *testing.T) {
		for _, tt := range []struct {
			name          string
			node          *bitmapIndexedNode[string, int]
			key           string
			value         int
			hash          uint64
			shift         uint
			expectedAdded bool
			expected      node[string, int]
		}{
			{
				name:          "set in empty node",
				node:          &bitmapIndexedNode[string, int]{},
				key:           "00001",
				value:         100,
				hash:          0b00001,
				shift:         0,
				expectedAdded: true,
				expected: &bitmapIndexedNode[string, int]{
					datamap: 0b00010,
					keys:    []string{"00001"},
					values:  []int{100},
				},
			},
			{
				name: "update existing key",
				node: &bitmapIndexedNode[string, int]{
					datamap: 0b00010,
					keys:    []string{"00001"},
					values:  []int{100},
				},
				key:           "00001",
				value:         200,
				hash:          0b00001,
				shift:         0,
				expectedAdded: false,
				expected: &bitmapIndexedNode[string, int]{
					datamap: 0b00010,
					keys:    []string{"00001"},
					values:  []int{200},
				},
			},
			{
				name: "collision creates child bit indexed node",
				node: &bitmapIndexedNode[string, int]{
					datamap: 0b00010,
					keys:    []string{"00001"},
					values:  []int{100},
				},
				key:           "0000100001",
				value:         200,
				hash:          0b0000100001,
				shift:         0,
				expectedAdded: true,
				expected: &bitmapIndexedNode[string, int]{
					nodemap: 0b00010,
					nodes: []node[string, int]{
						&bitmapIndexedNode[string, int]{
							datamap: 0b00011,
							keys:    []string{"00001", "0000100001"},
							values:  []int{100, 200},
						},
					},
				},
			},
			{
				name: "collision creates child collision node",
				node: &bitmapIndexedNode[string, int]{
					datamap: 0b00010,
					keys:    []string{"00001" + strings.Repeat("00000", 11)},
					values:  []int{100},
				},
				key:           "0000000001" + strings.Repeat("00000", 11),
				value:         200,
				hash:          0b0000000001 << (5 * 11),
				shift:         5 * 11,
				expectedAdded: true,
				expected: &bitmapIndexedNode[string, int]{
					nodemap: 0b00010,
					nodes: []node[string, int]{
						&bitmapIndexedNode[string, int]{
							nodemap: 0b00001,
							nodes: []node[string, int]{
								&collisionNode[string, int]{
									keys: []string{
										"0000000001" + strings.Repeat("00000", 11),
										"00001" + strings.Repeat("00000", 11),
									},
									values: []int{200, 100},
								},
							},
						},
					},
				},
			},
		} {
			t.Run(tt.name, func(t *testing.T) {
				result, added := tt.node.set(tt.key, tt.value, tt.hash, tt.shift, testHashFunc)
				if added != tt.expectedAdded {
					t.Errorf("set() added = %v, expected %v", added, tt.expectedAdded)
				}
				if !equalNode(result, tt.expected) {
					t.Errorf("set() result node not as expected\nactual:\n%s\nexpected:\n%s", result, tt.expected)
				}
			})
		}
	})

	t.Run("delete", func(t *testing.T) {
		for _, tt := range []struct {
			name            string
			node            *bitmapIndexedNode[string, int]
			key             string
			hash            uint64
			shift           uint
			expectedDeleted bool
			expected        node[string, int]
		}{
			{
				name:            "delete from empty node",
				node:            &bitmapIndexedNode[string, int]{},
				key:             "00001",
				hash:            0b00001,
				shift:           0,
				expectedDeleted: false,
				expected:        &bitmapIndexedNode[string, int]{},
			},
			{
				name: "delete only key returns nil",
				node: &bitmapIndexedNode[string, int]{
					datamap: 0b00010,
					keys:    []string{"00001"},
					values:  []int{100},
				},
				key:             "00001",
				hash:            0b00001,
				shift:           0,
				expectedDeleted: true,
				expected:        nil,
			},
			{
				name: "delete one of multiple keys",
				node: &bitmapIndexedNode[string, int]{
					datamap: 0b00111, // bits 0, 1, 2 set
					keys:    []string{"00000", "00001", "00010"},
					values:  []int{10, 20, 30},
				},
				key:             "00001",
				hash:            0b00001,
				shift:           0,
				expectedDeleted: true,
				expected: &bitmapIndexedNode[string, int]{
					datamap: 0b00101,
					keys:    []string{"00000", "00010"},
					values:  []int{10, 30},
				},
			},
			{
				name: "collapse single-entry child bitmap indexed node",
				node: &bitmapIndexedNode[string, int]{
					nodemap: 0b00001,
					nodes: []node[string, int]{
						&bitmapIndexedNode[string, int]{
							datamap: 0b00110,
							keys:    []string{"0000100000", "0000000010"},
							values:  []int{100, 200},
						},
					},
				},
				key:             "0000100000",
				hash:            0b0000100000,
				shift:           0,
				expectedDeleted: true,
				expected: &bitmapIndexedNode[string, int]{
					datamap: 0b00001,
					keys:    []string{"0000000010"},
					values:  []int{200},
				},
			},
			{
				name: "collapse single-entry child collision node",
				node: &bitmapIndexedNode[string, int]{
					nodemap: 0b00010,
					nodes: []node[string, int]{
						&bitmapIndexedNode[string, int]{
							nodemap: 0b00001,
							datamap: 0b00010,
							keys:    []string{"0000100001" + strings.Repeat("00000", 11)},
							values:  []int{100},
							nodes: []node[string, int]{
								&collisionNode[string, int]{
									keys: []string{
										"0000000001" + strings.Repeat("00000", 11),
										"00001" + strings.Repeat("00000", 11),
									},
									values: []int{200, 100},
								},
							},
						},
					},
				},
				key:             "00001" + strings.Repeat("00000", 11),
				hash:            0b00001 << (5 * 11),
				shift:           5 * 11,
				expectedDeleted: true,
				expected: &bitmapIndexedNode[string, int]{
					nodemap: 0b00010,
					nodes: []node[string, int]{
						&bitmapIndexedNode[string, int]{
							datamap: 0b00011,
							keys: []string{
								"0000000001" + strings.Repeat("00000", 11),
								"0000100001" + strings.Repeat("00000", 11),
							},
							values: []int{200, 100},
						},
					},
				},
			},
		} {
			t.Run(tt.name, func(t *testing.T) {
				result, deleted := tt.node.del(tt.key, tt.hash, tt.shift)
				if deleted != tt.expectedDeleted {
					t.Errorf("del() deleted = %v, expected %v", deleted, tt.expectedDeleted)
				}
				if !equalNode(result, tt.expected) {
					t.Errorf("set() result node not as expected\nactual:\n%s\nexpected:\n%s", result, tt.expected)
				}
			})
		}
	})
}

func TestCollisionNode(t *testing.T) {
	t.Run("get", func(t *testing.T) {
		for _, tt := range []struct {
			name       string
			node       *collisionNode[string, int]
			key        string
			expected   int
			expectedOk bool
		}{
			{
				name: "empty node",
				node: &collisionNode[string, int]{
					keys:   []string{},
					values: []int{},
				},
				key:        "00001",
				expected:   0,
				expectedOk: false,
			},
			{
				name: "single key present",
				node: &collisionNode[string, int]{
					keys:   []string{"00001"},
					values: []int{100},
				},
				key:        "00001",
				expected:   100,
				expectedOk: true,
			},
			{
				name: "multiple keys present",
				node: &collisionNode[string, int]{
					keys:   []string{"00000", "00010", "00100"},
					values: []int{10, 30, 50},
				},
				key:        "00010",
				expected:   30,
				expectedOk: true,
			},
			{
				name: "not found",
				node: &collisionNode[string, int]{
					keys:   []string{"00000", "00010", "00100"},
					values: []int{10, 30, 50},
				},
				key:        "00011",
				expected:   0,
				expectedOk: false,
			},
		} {
			t.Run(tt.name, func(t *testing.T) {
				got, ok := tt.node.get(tt.key, 0, 0) // hash and shift not used
				if ok != tt.expectedOk {
					t.Errorf("get() found = %v, expected %v", ok, tt.expectedOk)
				}
				if got != tt.expected {
					t.Errorf("get() value = %d, expected %d", got, tt.expected)
				}
			})
		}
	})

	t.Run("set", func(t *testing.T) {
		for _, tt := range []struct {
			name          string
			node          *collisionNode[string, int]
			key           string
			value         int
			expectedAdded bool
			expected      node[string, int]
		}{
			{
				name: "update existing key",
				node: &collisionNode[string, int]{
					keys:   []string{"00001", "00010"},
					values: []int{100, 200},
				},
				key:           "00010",
				value:         250,
				expectedAdded: false,
				expected: &collisionNode[string, int]{
					keys:   []string{"00001", "00010"},
					values: []int{100, 250},
				},
			},
			{
				name: "add new key to existing",
				node: &collisionNode[string, int]{
					keys:   []string{"00001", "00100"},
					values: []int{100, 200},
				},
				key:           "00010",
				value:         300,
				expectedAdded: true,
				expected: &collisionNode[string, int]{
					keys:   []string{"00001", "00010", "00100"},
					values: []int{100, 300, 200},
				},
			},
		} {
			t.Run(tt.name, func(t *testing.T) {
				result, added := tt.node.set(tt.key, tt.value, 0, 0, testHashFunc) // hash and shift not used
				if added != tt.expectedAdded {
					t.Errorf("set() added = %v, expected %v", added, tt.expectedAdded)
				}
				if !equalNode(result, tt.expected) {
					t.Errorf("set() result node not as expected\nactual:\n%s\nexpected:\n%s", result, tt.expected)
				}
			})
		}
	})

	t.Run("delete", func(t *testing.T) {
		for _, tt := range []struct {
			name            string
			node            *collisionNode[string, int]
			key             string
			expectedDeleted bool
			expected        node[string, int]
		}{
			{
				name: "delete non-existent key",
				node: &collisionNode[string, int]{
					keys:   []string{"00001", "00010"},
					values: []int{100, 200},
				},
				key:             "00100",
				expectedDeleted: false,
				expected: &collisionNode[string, int]{
					keys:   []string{"00001", "00010"},
					values: []int{100, 200},
				},
			},
			{
				name: "delete from many keys",
				node: &collisionNode[string, int]{
					keys:   []string{"00001", "00010", "00100"},
					values: []int{100, 200, 300},
				},
				key:             "00010",
				expectedDeleted: true,
				expected: &collisionNode[string, int]{
					keys:   []string{"00001", "00100"},
					values: []int{100, 300},
				},
			},
			{
				name: "delete from two keys converts to bitmap",
				node: &collisionNode[string, int]{
					keys:   []string{"00001", "00010"},
					values: []int{100, 200},
				},
				key:             "00010",
				expectedDeleted: true,
				expected: &bitmapIndexedNode[string, int]{
					datamap: 0b00001,
					keys:    []string{"00001"},
					values:  []int{100},
				},
			},
		} {
			t.Run(tt.name, func(t *testing.T) {
				result, deleted := tt.node.del(tt.key, 0, 0) // hash and shift not used
				if deleted != tt.expectedDeleted {
					t.Errorf("del() deleted = %v, expected %v", deleted, tt.expectedDeleted)
				}
				if !equalNode(result, tt.expected) {
					t.Errorf("del() result node not as expected\nactual:\n%s\nexpected:\n%s", result, tt.expected)
				}
			})
		}
	})
}

func TestPopcount(t *testing.T) {
	tests := []struct {
		input    uint32
		expected int
	}{
		{0x00000000, 0},
		{0x00000001, 1},
		{0x00000003, 2},
		{0x00000007, 3},
		{0x0000000F, 4},
		{0x000000FF, 8},
		{0x0000FFFF, 16},
		{0xFFFFFFFF, 32},
		{0x55555555, 16}, // alternating bits
		{0xAAAAAAAA, 16}, // alternating bits (opposite)
		{0x80000000, 1},  // highest bit only
		{0x00010000, 1},  // middle bit
	}

	for _, tt := range tests {
		result := popcount(tt.input)
		if result != tt.expected {
			t.Errorf("popcount(0x%08X) = %d, expected %d", tt.input, result, tt.expected)
		}
	}
}
