package champ_test

import (
	"fmt"

	"github.com/shota3506/go-champ"
)

func ExampleMap() {
	m := champ.New[string, int]()

	m = m.Set("apple", 5)
	m = m.Set("banana", 3)
	m = m.Set("cherry", 8)

	// Get a value
	if value, ok := m.Get("banana"); ok {
		fmt.Printf("banana: %d\n", value)
	}

	// Update a value
	m = m.Set("apple", 10)

	// Delete a key
	m = m.Delete("cherry")

	// Check the size
	fmt.Printf("Size: %d\n", m.Len())

	// Output:
	// banana: 3
	// Size: 2
}
