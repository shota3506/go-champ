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
BenchmarkMapGet/size_10-8               59344192                19.95 ns/op            0 B/op          0 allocs/op
BenchmarkMapGet/size_100-8              61596108                20.32 ns/op            0 B/op          0 allocs/op
BenchmarkMapGet/size_1000-8             40288735                30.02 ns/op            0 B/op          0 allocs/op
BenchmarkMapGet/size_10000-8            29187522                41.41 ns/op            0 B/op          0 allocs/op
BenchmarkMapGet/size_100000-8           20657718                59.40 ns/op            0 B/op          0 allocs/op
BenchmarkMapGet/size_1000000-8           4170073               278.3 ns/op             0 B/op          0 allocs/op
BenchmarkMapSet/update/size_10-8        10101579               120.5 ns/op           283 B/op          6 allocs/op
BenchmarkMapSet/update/size_100-8        6738126               177.8 ns/op           695 B/op          6 allocs/op
BenchmarkMapSet/update/size_1000-8               4709404               255.2 ns/op          1008 B/op          8 allocs/op
BenchmarkMapSet/update/size_10000-8              2676663               451.9 ns/op          1491 B/op          9 allocs/op
BenchmarkMapSet/update/size_100000-8             1543225               808.5 ns/op          1883 B/op         11 allocs/op
BenchmarkMapSet/update/size_1000000-8            1000000              1582 ns/op            2187 B/op         12 allocs/op
BenchmarkMapSet/insert/size_10-8                 1306065               981.2 ns/op          2176 B/op         12 allocs/op
BenchmarkMapSet/insert/size_100-8                1330926               976.0 ns/op          2179 B/op         12 allocs/op
BenchmarkMapSet/insert/size_1000-8               1300101               981.4 ns/op          2176 B/op         12 allocs/op
BenchmarkMapSet/insert/size_10000-8              1304326               976.2 ns/op          2183 B/op         12 allocs/op
BenchmarkMapSet/insert/size_100000-8             1216014              1013 ns/op            2215 B/op         12 allocs/op
BenchmarkMapSet/insert/size_1000000-8             936553              1312 ns/op            2338 B/op         13 allocs/op
BenchmarkMapDelete/size_10-8                     7906136               148.5 ns/op           287 B/op          7 allocs/op
BenchmarkMapDelete/size_100-8                    6600783               181.2 ns/op           673 B/op          6 allocs/op
BenchmarkMapDelete/size_1000-8                   4256062               282.7 ns/op          1073 B/op          7 allocs/op
BenchmarkMapDelete/size_10000-8                  2536194               473.8 ns/op          1491 B/op          8 allocs/op
BenchmarkMapDelete/size_100000-8                 1563249               767.0 ns/op          1865 B/op         10 allocs/op
BenchmarkMapDelete/size_1000000-8                 841875              1437 ns/op            2253 B/op         12 allocs/op
PASS
ok      github.com/shota3506/go-champ   49.664s
```

</details>

## License

This project is licensed under the MIT License.
