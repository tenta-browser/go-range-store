Go Range Store
==============

[![Build Status](https://travis-ci.org/tenta-browser/go-range-store.svg?branch=master)](https://travis-ci.org/tenta-browser/go-range-store/builds)
[![GoDoc](https://godoc.org/github.com/tenta-browser/go-range-store?status.svg)](https://godoc.org/github.com/tenta-browser/go-range-store)
[![Go Report Card](https://goreportcard.com/badge/github.com/tenta-browser/go-range-store)](https://goreportcard.com/report/github.com/tenta-browser/go-range-store)

Range store provides a simple datastructure providing efficient storage of a single value to many (consecutive) keys. Inspired by wanting to write:

```go
store := make(map[uint64]interface{})
store[10:20] = "hello"
greet := store[15]
// Greet is "hello"
```

Contact: developer@tenta.io

Installation
============

1. `go get github.com/tenta-browser/go-range-store`

Usage
=====

The range store provides a compact and efficient tree based method of storing a single value associated with a range of keys. You may
either use the provided `DefaultRangedValue` to store data. In addition, a `Ranged` interface is provided allowing for storage of
arbitrary types. For example, using US zip codes (postal codes), to map zip codes to a regional office:

```go
items := make([]Ranged,0)
items = append(items, DefaultRangedValue{0, 25505, "New York"})
items = append(items, DefaultRangedValue(25506, 67890, "Chicago"})
items = append(items, DefaultRangedValue(67891, 89000, "Phoenix"})
items = append(items, DefaultRangedValue(89001, 99999, "Los Angeles"})

n, err := NewRangeStoreFromSorted(items)
// Check error

city := n.RangeSearch(85716)

// City is "Phoenix"
```

The range store must be constructed with a continuously inscreasing set of non-negative integers which don't overlap and contain
no discontinuities. To simplify operation when using values where only weights matter, and not explicit ranges, a `Weighted` interface
and `DefaultWeightedValue` are provided. For example, so select fairly among servers with different weights:

```go
items := make([]Weighted, 0)
items = append(items, DefaultWeightedValue{10, "a.example.com"})
items = append(items, DefaultWeightedValue(10, "b.example.com")}
items = append(items, DefaultWeightedValue(20, "c.example.com")}

n, err := NewRangeStoreFromWeighted(items)
// Check error

p := rand.Intn(40)
server := n.RangeSearch(uint64(p))

// Server has a 25% chance of being a or b and a 50% chance of being c
```

Performance
===========

We wrote it to be fast and we use it in production. The tests include a totally arbitrary benchmark against a very naive implementation using
a raw map with an entry for every value in the range. Unsurprisingly, for large ranges, the range store massively outperforms the map. However,
even for small ranges, the range store still outperforms the map, as indexing requires only fast integer operations. In addition, the range store
always outperforms the map on construction:

```
Benchmark_NewNodeSorted_Small-4         10000000               219 ns/op
Benchmark_NewMapSorted_Small-4            300000              5389 ns/op
Benchmark_NewNodeSorted_Large-4         10000000               224 ns/op
Benchmark_NewMapSorted_Large-4                 5         203757320 ns/op
Benchmark_RangeSearch_Node-4            20000000                63 ns/op
Benchmark_RangeSearch_Map-4             10000000               186 ns/op
```

License
=======

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

For any questions, please contact developer@tenta.io

Contributing
============

We welcome contributions, feedback and plain old complaining. Feel free to open
an issue or shoot us a message to developer@tenta.io. If you'd like to contribute,
please open a pull request and send us an email to sign a contributor agreement.

About Tenta
===========

This range store library is brought to you by Team Tenta. Tenta is your [private, encrypted browser](https://tenta.com) that protects your data instead of selling. We're building a next-generation browser that combines all the privacy tools you need, including built-in OpenVPN. Everything is encrypted by default. That means your bookmarks, saved tabs, web history, web traffic, downloaded files, IP address and DNS. A truly incognito browser that's fast and easy.
