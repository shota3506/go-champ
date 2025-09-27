package champ

import (
	cryptorand "crypto/rand"
	"math/rand/v2"
	"strconv"
	"strings"
	"testing"
)

func FuzzMap(f *testing.F) {
	f.Add([]byte{byte(0)}) // set
	f.Add([]byte{byte(1)}) // get
	f.Add([]byte{byte(2)}) // delete
	f.Add([]byte(strings.Repeat("very long operation sequence source for deep testing", 100)))

	f.Fuzz(func(t *testing.T, ops []byte) {
		var seed [32]byte
		if _, err := cryptorand.Read(seed[:]); err != nil {
			t.Fatal(err)
		}
		r := rand.New(rand.NewChaCha8(seed))

		reference := make(map[string]int)
		m := New[string, int]()

		for _, op := range ops {
			key := "key" + strconv.FormatInt(r.Int64N(1<<16), 10)
			value := r.Int()

			switch op % 3 {
			case 0: // Set
				m = m.Set(key, value)
				reference[key] = value

				got, ok := m.Get(key)
				if !ok {
					t.Errorf("After Set(%q, %d), Get returned false", key, value)
				}
				if got != value {
					t.Errorf("After Set(%q, %d), Get returned %d", key, value, got)
				}
			case 1: // Get
				got, ok := m.Get(key)
				expectedVal, expectedOk := reference[key]

				if ok != expectedOk {
					t.Errorf("Get(%q) returned ok=%v, expected %v", key, ok, expectedOk)
				}
				if ok && got != expectedVal {
					t.Errorf("Get(%q) returned %d, expected %d", key, got, expectedVal)
				}
			case 2: // Delete
				m = m.Delete(key)
				delete(reference, key)

				if _, ok := m.Get(key); ok {
					t.Errorf("After Delete(%q), Get returned true", key)
				}
			}

			if m.Len() != len(reference) {
				t.Errorf("Map.Len() = %d, but reference map has %d entries", m.Len(), len(reference))
			}
		}
	})
}
