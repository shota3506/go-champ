package champ

import (
	"fmt"
	"testing"
)

type op func(t *testing.T, m *Map[string, int]) *Map[string, int]

func get(key string, expectedValue int, expectedOk bool) op {
	return func(t *testing.T, m *Map[string, int]) *Map[string, int] {
		t.Helper()

		v, ok := m.Get(key)
		if ok != expectedOk {
			t.Fatalf("Get(%q) expected ok=%v, actual %v", key, expectedOk, ok)
		}
		if v != expectedValue {
			t.Fatalf("Get(%q) expected value=%d, actual %d", key, expectedValue, v)
		}
		return m
	}
}

func set(key string, value int) op {
	return func(_ *testing.T, m *Map[string, int]) *Map[string, int] {
		return m.Set(key, value)
	}
}

func del(key string) op {
	return func(_ *testing.T, m *Map[string, int]) *Map[string, int] {
		return m.Delete(key)
	}
}

func length(expected int) op {
	return func(t *testing.T, m *Map[string, int]) *Map[string, int] {
		t.Helper()

		if m.Len() != expected {
			t.Fatalf("Len() expected %d, actual %d", expected, m.Len())
		}
		return m
	}
}

func TestMap(t *testing.T) {
	t.Run("basic operations", func(t *testing.T) {
		ops := []op{
			length(0),
			get("key1", 0, false),
			set("key1", 10),
			length(1),
			get("key1", 10, true),
			set("key1", 20),
			length(1),
			get("key1", 20, true),
			set("key2", 30),
			length(2),
			set("key3", 40),
			length(3),
			del("key2"),
			length(2),
			get("key2", 0, false),
			del(""),
			length(2),
			del("key3"),
			del("key1"),
			length(0),
		}

		m := New[string, int]()
		for _, op := range ops {
			m = op(t, m)
		}
	})

	t.Run("large map", func(t *testing.T) {
		const n = 2048

		m := New[string, int]()
		for i := range n {
			m = m.Set(fmt.Sprintf("key%d", i), i)
		}

		// verify
		m = length(n)(t, m)
		for i := range n {
			m = get(fmt.Sprintf("key%d", i), i, true)(t, m)
		}

		// delete half
		for i := 0; i < n; i += 2 {
			m = m.Delete(fmt.Sprintf("key%d", i))
		}

		// verify
		m = length(n/2)(t, m)
		for i := range n {
			if i%2 == 0 {
				m = get(fmt.Sprintf("key%d", i), 0, false)(t, m)
			} else {
				m = get(fmt.Sprintf("key%d", i), i, true)(t, m)
			}
		}
	})

	t.Run("immutability", func(t *testing.T) {
		step := []struct {
			op       op
			expected *Map[string, int]
		}{
			{
				op:       set("key1", 10),
				expected: New[string, int]().Set("key1", 10),
			},
			{
				op:       set("key2", 20),
				expected: New[string, int]().Set("key1", 10).Set("key2", 20),
			},
			{
				op:       set("key1", 100),
				expected: New[string, int]().Set("key1", 100).Set("key2", 20),
			},
			{
				op:       set("key3", 30),
				expected: New[string, int]().Set("key1", 100).Set("key2", 20).Set("key3", 30),
			},
			{
				op:       del("key2"),
				expected: New[string, int]().Set("key1", 100).Set("key3", 30),
			},
			{
				op:       del("key3"),
				expected: New[string, int]().Set("key1", 100),
			},
		}

		assertFuncs := []func() (int, bool){}

		m := New[string, int]()
		for i, s := range step {
			m = s.op(t, m)
			m := m
			assertFuncs = append(assertFuncs, func() (int, bool) {
				return i, Equal(m, s.expected)
			})
		}

		for _, assert := range assertFuncs {
			i, ok := assert()
			if !ok {
				t.Errorf("step %d map does not match expected", i)
			}
		}
	})
}

func TestImmutability(t *testing.T) {
	m1 := New[string, int]()
	m1 = m1.Set("key1", 10)
	m2 := m1.Set("key2", 20)

	// m1 should not be affected by changes to m2
	if m1.Len() != 1 {
		t.Errorf("Original map should still have length 1, actual %d", m1.Len())
	}

	_, ok := m1.Get("key2")
	if ok {
		t.Error("Original map should not have key2")
	}

	// m2 should have both keys
	if m2.Len() != 2 {
		t.Errorf("New map should have length 2, actual %d", m2.Len())
	}

	v, ok := m2.Get("key1")
	if !ok || v != 10 {
		t.Error("New map should have key1 with value 10")
	}

	v, ok = m2.Get("key2")
	if !ok || v != 20 {
		t.Error("New map should have key2 with value 20")
	}
}

func TestMapAll(t *testing.T) {
	for _, tt := range []struct {
		name     string
		expected map[string]int
	}{
		{
			name:     "empty",
			expected: map[string]int{},
		},
		{
			name: "small map",
			expected: map[string]int{
				"key1": 1,
				"key2": 2,
				"key3": 3,
			},
		},
		{
			name: "large map",
			expected: func() map[string]int {
				const n = 2048
				m := make(map[string]int, n)
				for i := range n {
					m[fmt.Sprintf("key%d", i)] = i
				}
				return m
			}(),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			m := New[string, int]()
			for k, v := range tt.expected {
				m = m.Set(k, v)
			}

			for k, v := range m.All() {
				if tt.expected[k] != v {
					t.Errorf("All() returned unexpected value for key %s: expected %d, actual %d", k, tt.expected[k], v)
				}
				delete(tt.expected, k)
			}
			if len(tt.expected) != 0 {
				t.Error("All() did not return all expected key-value pairs")
			}
		})
	}
}

func TestEqual(t *testing.T) {
	for _, tt := range []struct {
		name     string
		maps     func() (*Map[string, int], *Map[string, int])
		expected bool
	}{
		{
			name: "empty",
			maps: func() (*Map[string, int], *Map[string, int]) {
				return New[string, int](), New[string, int]()
			},
			expected: true,
		},
		{
			name: "same items in same order",
			maps: func() (*Map[string, int], *Map[string, int]) {
				m1 := New[string, int]().Set("a", 1).Set("b", 2).Set("c", 3)
				m2 := New[string, int]().Set("a", 1).Set("b", 2).Set("c", 3)
				return m1, m2
			},
			expected: true,
		},
		{
			name: "same items in different order",
			maps: func() (*Map[string, int], *Map[string, int]) {
				m1 := New[string, int]().Set("a", 1).Set("b", 2).Set("c", 3)
				m2 := New[string, int]().Set("c", 3).Set("b", 2).Set("a", 1)
				return m1, m2
			},
			expected: true,
		},
		{
			name: "same large maps with mutations",
			maps: func() (*Map[string, int], *Map[string, int]) {
				const n = 2048
				m1 := New[string, int]()
				for i := 0; i < n; i += 2 {
					m1 = m1.Set(fmt.Sprintf("key%d", i), i)
				}
				m2 := m1
				for i := 1; i < n; i += 2 {
					m2 = m2.Set(fmt.Sprintf("key%d", i), i)
				}
				for i := 1; i < n; i += 2 {
					m2 = m2.Delete(fmt.Sprintf("key%d", i))
				}
				return m1, m2
			},
			expected: true,
		},
		{
			name: "same items with different operations",
			maps: func() (*Map[string, int], *Map[string, int]) {
				m1 := New[string, int]().Set("a", 1).Set("c", 3)
				m2 := New[string, int]().Set("a", 1).Set("b", 2).Set("c", 4).Delete("b").Set("c", 3)
				return m1, m2
			},
			expected: true,
		},
		{
			name: "different sizes",
			maps: func() (*Map[string, int], *Map[string, int]) {
				m1 := New[string, int]().Set("a", 1).Set("b", 2)
				m2 := New[string, int]().Set("a", 1).Set("b", 2).Set("c", 3)
				return m1, m2
			},
			expected: false,
		},
		{
			name: "different values",
			maps: func() (*Map[string, int], *Map[string, int]) {
				m1 := New[string, int]().Set("a", 1).Set("b", 2).Set("c", 3)
				m2 := New[string, int]().Set("a", 1).Set("b", 20).Set("c", 3)
				return m1, m2
			},
			expected: false,
		},
		{
			name: "different keys",
			maps: func() (*Map[string, int], *Map[string, int]) {
				m1 := New[string, int]().Set("a", 1).Set("b", 2).Set("c", 3)
				m2 := New[string, int]().Set("a", 1).Set("b", 2).Set("d", 3)
				return m1, m2
			},
			expected: false,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			m1, m2 := tt.maps()
			if got := Equal(m1, m2); got != tt.expected {
				t.Errorf("Equal() = %v, expected %v", got, tt.expected)
			}
		})
	}
}
