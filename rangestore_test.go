/**
 * Go Range Store
 *
 *    Copyright 2017 Tenta, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * For any questions, please contact developer@tenta.io
 *
 * rangestore_test.go: Tests on the core range store
 */

package rangestore

import (
	"math/rand"
	"reflect"
	"testing"
)

func TestRangeStoreFromSorted_Basic(t *testing.T) {
	items := make([]Ranged, 0)

	items = append(items, DefaultRangedValue{0, 9, "A"})
	items = append(items, DefaultRangedValue{10, 19, "B"})
	items = append(items, DefaultRangedValue{20, 29, "C"})

	n, err := NewRangeStoreFromSorted(items)

	if err != nil {
		t.Fatalf("Error while constructing range store: %s", err.Error())
	}

	// -B [max: 19]
	//  |-A [max: 9]
	//  !-C [max: 29]
	if n.value != "B" {
		t.Fatalf("Expected B at the root")
	}
	if n.max != 19 {
		t.Fatalf("Expected 19 max at the root")
	}
	if n.left.value != "A" {
		t.Fatalf("Expected A as the right child")
	}
	if n.left.max != 9 {
		t.Fatalf("Expected 9 max as the right child")
	}
	if n.right.value != "C" {
		t.Fatalf("Expected C as the right child")
	}
	if n.right.max != 29 {
		t.Fatalf("Expected 29 max as the right child")
	}

	if n.left.left != nil {
		t.Fatalf("Expected children to have null leaves")
	}
	if n.left.right != nil {
		t.Fatalf("Expected children to have null leaves")
	}
	if n.right.left != nil {
		t.Fatalf("Expected children to have null leaves")
	}
	if n.right.right != nil {
		t.Fatalf("Expected children to have null leaves")
	}
}

func TestRangeStoreFromSorted_Overflow(t *testing.T) {
	items := make([]Ranged, 0)

	items = append(items, DefaultRangedValue{0, (1 << 63), "A"})
	items = append(items, DefaultRangedValue{(1 << 63) + 1, 19, "B"})
	items = append(items, DefaultRangedValue{20, 29, "C"})

	_, err := NewRangeStoreFromSorted(items)

	if err == nil {
		t.Fatalf("Expecting integer overflow error and got none")
	}
	if reflect.TypeOf(err).Name() != reflect.TypeOf(ErrUnsignedIntegerOverflow{}).Name() {
		t.Fatalf("Expecting an ErrUnsignedIntegerOverflow, but got something else")
	}
	msg := err.Error()
	if msg != "Overflow adding 9223372036854775809 + 9223372036854775827" {
		t.Fatalf("Wrong error message: %s", msg)
	}
}

func TestRangeStoreFromSorted_Overlap(t *testing.T) {
	items := make([]Ranged, 0)

	items = append(items, DefaultRangedValue{0, 10, "A"})
	items = append(items, DefaultRangedValue{9, 19, "B"})
	items = append(items, DefaultRangedValue{20, 29, "C"})

	_, err := NewRangeStoreFromSorted(items)

	if err == nil {
		t.Fatalf("Expecting overlap error and got none")
	}
	if reflect.TypeOf(err).Name() != reflect.TypeOf(ErrOverlap{}).Name() {
		t.Fatalf("Expecting an ErrOverlap, but got something else")
	}
	msg := err.Error()
	if msg != "Overlap detected between 10 -> 9" {
		t.Fatalf("Wrong error message: %s", msg)
	}
}

func TestNode_RangeSearch(t *testing.T) {
	items := make([]Ranged, 0)

	items = append(items, DefaultRangedValue{0, 9, "A"})
	items = append(items, DefaultRangedValue{10, 19, "B"})
	items = append(items, DefaultRangedValue{20, 29, "C"})

	n, err := NewRangeStoreFromSorted(items)

	if err != nil {
		t.Fatalf("Error while constructing range store: %s", err.Error())
	}

	a0, err := n.RangeSearch(0)
	if err != nil {
		t.Fatalf("Got an error while searching: %s", err.Error())
	}
	if a0 != "A" {
		t.Fatalf("Got invalid value back %s [%s]", a0, "A")
	}
	a3, err := n.RangeSearch(3)
	if err != nil {
		t.Fatalf("Got an error while searching: %s", err.Error())
	}
	if a3 != "A" {
		t.Fatalf("Got invalid value back %s [%s]", a3, "A")
	}
	a9, err := n.RangeSearch(9)
	if err != nil {
		t.Fatalf("Got an error while searching: %s", err.Error())
	}
	if a0 != "A" {
		t.Fatalf("Got invalid value back %s [%s]", a9, "A")
	}

	b0, err := n.RangeSearch(10)
	if err != nil {
		t.Fatalf("Got an error while searching: %s", err.Error())
	}
	if b0 != "B" {
		t.Fatalf("Got invalid value back %s [%s]", b0, "B")
	}
	b5, err := n.RangeSearch(15)
	if err != nil {
		t.Fatalf("Got an error while searching: %s", err.Error())
	}
	if b5 != "B" {
		t.Fatalf("Got invalid value back %s [%s]", b5, "B")
	}
	b9, err := n.RangeSearch(19)
	if err != nil {
		t.Fatalf("Got an error while searching: %s", err.Error())
	}
	if b9 != "B" {
		t.Fatalf("Got invalid value back %s [%s]", b9, "B")
	}

	c0, err := n.RangeSearch(20)
	if err != nil {
		t.Fatalf("Got an error while searching: %s", err.Error())
	}
	if c0 != "C" {
		t.Fatalf("Got invalid value back %s [%s]", c0, "C")
	}
	c7, err := n.RangeSearch(27)
	if err != nil {
		t.Fatalf("Got an error while searching: %s", err.Error())
	}
	if c7 != "C" {
		t.Fatalf("Got invalid value back %s [%s]", c7, "C")
	}
	c9, err := n.RangeSearch(29)
	if err != nil {
		t.Fatalf("Got an error while searching: %s", err.Error())
	}
	if c9 != "C" {
		t.Fatalf("Got invalid value back %s [%s]", c9, "C")
	}

	_, err = n.RangeSearch(30)
	if err == nil {
		t.Fatalf("Expected an error while performing an out of range search, got nothing")
	}
	if reflect.TypeOf(err).Name() != reflect.TypeOf(ErrOutOfRange{}).Name() {
		t.Fatalf("Expecting an ErrOutOfRange, but got something else")
	}
	msg := err.Error()
	if msg != "Value 30 is out of range" {
		t.Fatalf("Wrong error message: %s", msg)
	}
}

func TestRangeStoreFromSorted_Lots(t *testing.T) {
	items := make([]Ranged, 0)

	items = append(items, DefaultRangedValue{0, 9, "A"})
	items = append(items, DefaultRangedValue{10, 19, "B"})
	items = append(items, DefaultRangedValue{20, 29, "C"})
	items = append(items, DefaultRangedValue{30, 39, "D"})
	items = append(items, DefaultRangedValue{40, 49, "E"})
	items = append(items, DefaultRangedValue{50, 59, "F"})
	items = append(items, DefaultRangedValue{60, 69, "G"})
	items = append(items, DefaultRangedValue{70, 79, "H"})
	items = append(items, DefaultRangedValue{80, 89, "I"})
	items = append(items, DefaultRangedValue{90, 99, "J"})

	_, err := NewRangeStoreFromSorted(items)

	if err != nil {
		t.Fatalf("Error while constructing range store: %s", err.Error())
	}
}

func TestRangeStoreFromSorted_LongTail(t *testing.T) {
	items := make([]Ranged, 0)

	items = append(items, DefaultRangedValue{0, 2, "A"})
	items = append(items, DefaultRangedValue{3, 5, "B"})
	items = append(items, DefaultRangedValue{6, 29, "C"})

	n, err := NewRangeStoreFromSorted(items)

	if err != nil {
		t.Fatalf("Error while constructing range store: %s", err.Error())
	}

	//-C [max: 29]
	// |-A [max: 2]
	// | !-B [max: 5]
	if n.value != "C" {
		t.Fatalf("Expected B at the root")
	}
	if n.max != 29 {
		t.Fatalf("Expected 19 max at the root")
	}
	if n.left.value != "A" {
		t.Fatalf("Expected A as the left child")
	}
	if n.left.max != 2 {
		t.Fatalf("Expected 2 max as the left child")
	}
	if n.right != nil {
		t.Fatalf("Expected nil right child")
	}
	if n.left.right.value != "B" {
		t.Fatalf("Expected B as the left-right grandchild")
	}
	if n.left.right.max != 5 {
		t.Fatalf("Expected 5 as the max left-right grandchild")
	}

	if n.left.left != nil {
		t.Fatalf("Expected children to have null leaves")
	}
	if n.left.right.left != nil {
		t.Fatalf("Expected children to have null leaves")
	}
	if n.left.right.right != nil {
		t.Fatalf("Expected children to have null leaves")
	}
}

func TestRangeStoreFromSorted_Discontinuity(t *testing.T) {
	items := make([]Ranged, 0)

	items = append(items, DefaultRangedValue{0, 9, "A"})
	items = append(items, DefaultRangedValue{11, 19, "B"})
	items = append(items, DefaultRangedValue{20, 29, "C"})

	_, err := NewRangeStoreFromSorted(items)

	if err == nil {
		t.Fatalf("Error while constructing range store: Expected an error, but none generated")
	}
	if reflect.TypeOf(err).Name() != reflect.TypeOf(ErrDiscontinuity{}).Name() {
		t.Fatalf("Expecting an ErrDiscontinuity, but got something else")
	}
	msg := err.Error()
	if msg != "Discontinuity detected from 9 -> 11" {
		t.Fatalf("Wrong error message: %s", msg)
	}
}

func TestRangeStoreFromSorted_Empty(t *testing.T) {
	items := make([]Ranged, 0)

	_, err := NewRangeStoreFromSorted(items)

	if err == nil {
		t.Fatalf("Error while constructing range store: Expected an error, but none generated")
	}
	if reflect.TypeOf(err).Name() != reflect.TypeOf(ErrEmptyInput{}).Name() {
		t.Fatalf("Expecting an ErrEmptyInput, but got something else")
	}
	msg := err.Error()
	if msg != "Input list is empty" {
		t.Fatalf("Wrong error message: %s", msg)
	}
}

func Benchmark_NewNodeSorted_Small(b *testing.B) {
	items := make([]Ranged, 0)
	items = append(items, DefaultRangedValue{0, 9, "A"})
	items = append(items, DefaultRangedValue{10, 19, "B"})
	items = append(items, DefaultRangedValue{20, 29, "C"})
	b.ResetTimer()
	for n := 0; n < b.N; n += 1 {
		_, err := NewRangeStoreFromSorted(items)
		if err != nil {
			b.Fatalf("Got an error while benchmarking: %s", err.Error())
		}
	}
}

func Benchmark_NewMapSorted_Small(b *testing.B) {
	items := make([]Ranged, 0)
	items = append(items, DefaultRangedValue{0, 9, "A"})
	items = append(items, DefaultRangedValue{10, 19, "B"})
	items = append(items, DefaultRangedValue{20, 29, "C"})
	b.ResetTimer()
	for n := 0; n < b.N; n += 1 {
		ret := make(map[uint64]interface{})
		for _, item := range items {
			for i := item.GetMin(); i <= item.GetMax(); i += 1 {
				ret[i] = item.GetValue()
			}
		}
		if _, ok := ret[0]; !ok {
			b.Fatalf("Failed to generate a valid map")
		}
	}
}

func Benchmark_NewNodeSorted_Large(b *testing.B) {
	items := make([]Ranged, 0)
	items = append(items, DefaultRangedValue{0, 199999, "A"})
	items = append(items, DefaultRangedValue{200000, 399999, "B"})
	items = append(items, DefaultRangedValue{400000, 599999, "C"})
	b.ResetTimer()
	for n := 0; n < b.N; n += 1 {
		_, err := NewRangeStoreFromSorted(items)
		if err != nil {
			b.Fatalf("Got an error while benchmarking: %s", err.Error())
		}
	}
}

func Benchmark_NewMapSorted_Large(b *testing.B) {
	items := make([]Ranged, 0)
	items = append(items, DefaultRangedValue{0, 199999, "A"})
	items = append(items, DefaultRangedValue{200000, 399999, "B"})
	items = append(items, DefaultRangedValue{400000, 599999, "C"})
	b.ResetTimer()
	for n := 0; n < b.N; n += 1 {
		ret := make(map[uint64]interface{})
		for _, item := range items {
			for i := item.GetMin(); i <= item.GetMax(); i += 1 {
				ret[i] = item.GetValue()
			}
		}
		if _, ok := ret[0]; !ok {
			b.Fatalf("Failed to generate a valid map")
		}
	}
}

func Benchmark_RangeSearch_Node(b *testing.B) {
	items := make([]Ranged, 0)
	items = append(items, DefaultRangedValue{0, 199999, "A"})
	items = append(items, DefaultRangedValue{200000, 399999, "B"})
	items = append(items, DefaultRangedValue{400000, 600000, "C"})
	n, _ := NewRangeStoreFromSorted(items)
	b.ResetTimer()
	for i := 0; i < b.N; i += 1 {
		v := uint64(rand.Int() % 600000)
		e, err := n.RangeSearch(v)
		if v >= 0 && v < 200000 && e != "A" {
			b.Fatalf("Wrong value %d -> %s", v, e)
		}
		if v >= 200000 && v < 400000 && e != "B" {
			b.Fatalf("Wrong value %d -> %s", v, e)
		}
		if v >= 400000 && v < 600000 && e != "C" {
			b.Fatalf("Wrong value %d -> %s", v, e)
		}
		if err != nil {
			b.Fatalf("Got an error while searching: %s", err.Error())
		}
	}
}

func Benchmark_RangeSearch_Map(b *testing.B) {
	items := make([]Ranged, 0)
	items = append(items, DefaultRangedValue{0, 199999, "A"})
	items = append(items, DefaultRangedValue{200000, 399999, "B"})
	items = append(items, DefaultRangedValue{400000, 600000, "C"})
	ret := make(map[uint64]interface{})
	for _, item := range items {
		for i := item.GetMin(); i <= item.GetMax(); i += 1 {
			ret[i] = item.GetValue()
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i += 1 {
		v := uint64(rand.Int() % 600000)
		e, ok := ret[v]
		if !ok {
			b.Fatalf("Error while testing, unable to get in range value")
		}
		if v >= 0 && v < 200000 && e != "A" {
			b.Fatalf("Wrong value %d -> %s", v, e)
		}
		if v >= 200000 && v < 400000 && e != "B" {
			b.Fatalf("Wrong value %d -> %s", v, e)
		}
		if v >= 400000 && v < 600000 && e != "C" {
			b.Fatalf("Wrong value %d -> %s", v, e)
		}
	}
}

func TestNode_String(t *testing.T) {
	// Meh, print the string representation to get coverage on that code.
	items := make([]Ranged, 0)

	items = append(items, DefaultRangedValue{0, 9, "A"})
	items = append(items, DefaultRangedValue{10, 19, "B"})
	items = append(items, DefaultRangedValue{20, 29, "C"})

	n, err := NewRangeStoreFromSorted(items)

	if err != nil {
		t.Fatalf("Error while constructing range store: %s", err.Error())
	}

	R := `-B [max: 19]
 |-A [max: 9]
 !-C [max: 29]
`

	str := n.String()

	if str != R {
		t.Fatalf("Wrong string output form:\n%s\n%s", str, R)
	}
}
