package champ

import (
	"fmt"
	"math/rand/v2"
	"strconv"
	"testing"
)

func BenchmarkMapGet(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000, 100000, 1000000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			m := New[string, int]()
			keys := make([]string, size)
			for i := range size {
				key := strconv.FormatInt(int64(i), 2)
				keys[i] = key
				m = m.Set(key, i)
			}

			b.ResetTimer()

			for b.Loop() {
				key := keys[rand.IntN(size)]
				_, _ = m.Get(key)
			}
		})
	}
}

func BenchmarkMapSet(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000, 100000, 1000000}

	b.Run("update", func(b *testing.B) {
		for _, size := range sizes {
			b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
				m := New[string, int]()
				keys := make([]string, size)
				for i := range size {
					key := strconv.FormatInt(int64(i), 2)
					keys[i] = key
					m = m.Set(key, i)
				}

				b.ResetTimer()

				for i := range b.N {
					key := keys[rand.IntN(size)]
					m = m.Set(key, i)
				}
			})
		}
	})

	b.Run("insert", func(b *testing.B) {
		for _, size := range sizes {
			b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
				m := New[string, int]()
				for i := range size {
					key := strconv.FormatInt(int64(i), 2)
					m = m.Set(key, i)
				}

				newKeys := make([]string, b.N)
				for i := range b.N {
					key := strconv.FormatInt(int64(size+i), 2)
					newKeys[i] = key
				}

				b.ResetTimer()

				for i := range b.N {
					m = m.Set(newKeys[i], i)
				}
			})
		}
	})
}

func BenchmarkMapDelete(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000, 100000, 1000000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			m := New[string, int]()
			keys := make([]string, size)
			for i := range size {
				key := strconv.FormatInt(int64(i), 2)
				keys[i] = key
				m = m.Set(key, i)
			}

			b.ResetTimer()

			for b.Loop() {
				key := keys[rand.IntN(size)]
				_ = m.Delete(key)
			}
		})
	}
}
