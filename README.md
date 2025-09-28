# go-champ

[![Go Reference](https://pkg.go.dev/badge/github.com/shota3506/go-champ.svg)](https://pkg.go.dev/github.com/shota3506/go-champ)
[![Go Report Card](https://goreportcard.com/badge/github.com/shota3506/go-champ)](https://goreportcard.com/report/github.com/shota3506/go-champ)

A Go implementation of CHAMP (Compressed Hash-Array Mapped Prefix-tree), an efficient immutable map data structure.

## Overview

CHAMP is a persistent data structure that provides efficient immutable maps with structural sharing. It offers:
- **O(log₃₂ n)** time complexity for get, set, and delete operations
- **Immutable operations** - all modifications return a new map instance
- **Memory efficient** - uses structural sharing between versions

See the original paper for more details.

Michael J. Steindorfer and Jurgen J. Vinju. 2015. Optimizing hash-array mapped tries for fast and lean immutable JVM collections. SIGPLAN Not. 50, 10 (October 2015), 783–800. https://doi.org/10.1145/2858965.2814312

## Installation

```bash
go get github.com/shota3506/go-champ
```

## Usage

```go
package main

import (
	"fmt"

	"github.com/shota3506/go-champ"
)

func main() {
	m := champ.New[string, int]()

	// Set values (immutable, returns a new map)
	m = m.Set("key1", 100)
	m = m.Set("key2", 200)
	m = m.Set("key3", 300)

	// Get value
	if value, ok := m.Get("key1"); ok {
		fmt.Printf("key1: %d\n", value) // Output: key1: 100
	}

	// Delete a key (immutable, returns a new map)
	m = m.Delete("key2")

	// Check map size
	fmt.Printf("Size: %d\n", m.Len()) // Output: Size: 2

	// Iterate over all key-value pairs
	for key, value := range m.All() {
		fmt.Printf("%s: %d\n", key, value)
	}
}
```

## Performance

Map.Get: O(log₃₂ n)

Map.Set: O(log₃₂ n)

Map.Delete: O(log₃₂ n)

Map.Len: O(1)

Map.All: O(n) to iterate over all key-value pairs

<details>

<summary>Benchmark results</summary>

```
goos: darwin
goarch: arm64
pkg: github.com/shota3506/go-champ
cpu: Apple M3
BenchmarkMapGet/size_10-8               58748752                20.37 ns/op            0 B/op          0 allocs/op
lBenchmarkMapGet/size_100-8             47989282                24.85 ns/op            0 B/op          0 allocs/op
BenchmarkMapGet/size_1000-8             39153734                30.76 ns/op            0 B/op            0 allocs/op
BenchmarkMapGet/size_10000-8            29500065                40.40 ns/op            0 B/op            0 allocs/op
BenchmarkMapGet/size_100000-8           23263834                51.22 ns/op            0 B/op            0 allocs/op
BenchmarkMapGet/size_1000000-8           4214810               273.6 ns/op             0 B/op            0 allocs/op
BenchmarkMapSet/update/size_10-8        11112537               107.3 ns/op           302 B/op            5 allocs/op
BenchmarkMapSet/update/size_100-8        6494270               184.0 ns/op           695 B/op            7 allocs/op
BenchmarkMapSet/update/size_1000-8               4796413               251.8 ns/op     999 B/op          8 allocs/op
BenchmarkMapSet/update/size_10000-8              2642210               452.0 ns/op    1492 B/op          9 allocs/op
BenchmarkMapSet/update/size_100000-8             1558410               919.8 ns/op    1886 B/op         11 allocs/op
BenchmarkMapSet/update/size_1000000-8            1000000              1563 ns/op     2186 B/op          12 allocs/op
BenchmarkMapSet/insert/size_10-8                 1000000              1028 ns/op     2135 B/op          12 allocs/op
BenchmarkMapSet/insert/size_100-8                1000000              1059 ns/op     2135 B/op          12 allocs/op
BenchmarkMapSet/insert/size_1000-8               1000000              1065 ns/op     2137 B/op          12 allocs/op
BenchmarkMapSet/insert/size_10000-8              1000000              1094 ns/op     2145 B/op          12 allocs/op
BenchmarkMapSet/insert/size_100000-8             1000000              1081 ns/op     2192 B/op          12 allocs/op
BenchmarkMapSet/insert/size_1000000-8             835600              1472 ns/op     2332 B/op          13 allocs/op
BenchmarkMapDelete/size_10-8                    10169792               112.5 ns/op     310 B/op          4 allocs/op
BenchmarkMapDelete/size_100-8                    6187014               193.4 ns/op     680 B/op          6 allocs/op
BenchmarkMapDelete/size_1000-8                   4267792               280.7 ns/op    1058 B/op          8 allocs/op
BenchmarkMapDelete/size_10000-8                  2590080               465.0 ns/op    1492 B/op          8 allocs/op
BenchmarkMapDelete/size_100000-8                 1574355               762.6 ns/op    1869 B/op         10 allocs/op
BenchmarkMapDelete/size_1000000-8                1000000              1453 ns/op     2251 B/op          12 allocs/op
PASS
ok      github.com/shota3506/go-champ   44.873s
```

</details>

## License

This project is licensed under the MIT License.
