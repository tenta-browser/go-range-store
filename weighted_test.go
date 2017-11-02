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
 * weighted_test.go: Tests on the weighted wrapper
 */

package rangestore

import (
	"reflect"
	"testing"
)

func TestRangeStoreFromWeighted_Basic(t *testing.T) {
	vals := make([]Weighted, 0)
	vals = append(vals, &DefaultWeightedValue{Weight: 9, Value: "A"})
	vals = append(vals, &DefaultWeightedValue{Weight: 10, Value: "B"})
	vals = append(vals, &DefaultWeightedValue{Weight: 10, Value: "C"})

	n, err := NewRangeStoreFromWeighted(vals)

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

func TestRangeStoreFromWeighted_Empty(t *testing.T) {
	items := make([]Weighted, 0)

	n, err := NewRangeStoreFromWeighted(items)
	if n != nil {
		t.Fatal("Expected no return but got one")
	}
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

func TestRangeStoreFromWeighted_Overflow(t *testing.T) {
	items := make([]Weighted, 0)
	items = append(items, DefaultWeightedValue{1 << 63, "A"})
	items = append(items, DefaultWeightedValue{1 << 63, "B"})

	_, err := NewRangeStoreFromWeighted(items)

	if err == nil {
		t.Fatalf("Expecting integer overflow error and got none")
	}
	if reflect.TypeOf(err).Name() != reflect.TypeOf(ErrUnsignedIntegerOverflow{}).Name() {
		t.Fatalf("Expecting an ErrUnsignedIntegerOverflow, but got something else")
	}
	msg := err.Error()
	if msg != "Overflow adding 9223372036854775808 + 9223372036854775808" {
		t.Fatalf("Wrong error message: %s", msg)
	}
}
