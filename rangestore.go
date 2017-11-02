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
 * rangestore.go: Range Store Implementation
 */

package rangestore

import (
	"fmt"
)

type Node struct {
	max         uint64
	value       interface{}
	left, right *Node
}

type Weighted interface {
	GetWeight() uint64
	GetValue() interface{}
}
type DefaultWeightedValue struct {
	Weight uint64
	Value  interface{}
}

func (w DefaultWeightedValue) GetWeight() uint64 {
	return w.Weight
}
func (w DefaultWeightedValue) GetValue() interface{} {
	return w.Value
}

type Ranged interface {
	GetMin() uint64
	GetMax() uint64
	GetValue() interface{}
}
type DefaultRangedValue struct {
	min, max uint64
	value    interface{}
}

func (r DefaultRangedValue) GetMin() uint64 {
	return r.min
}
func (r DefaultRangedValue) GetMax() uint64 {
	return r.max
}
func (r DefaultRangedValue) GetValue() interface{} {
	return r.value
}

type ErrUnsignedIntegerOverflow struct {
	a, b uint64
}

func (ex ErrUnsignedIntegerOverflow) Error() string {
	return fmt.Sprintf("Overflow adding %d + %d", ex.a, ex.b)
}

type ErrDiscontinuity struct {
	x, y uint64
}

func (ex ErrDiscontinuity) Error() string {
	return fmt.Sprintf("Discontinuity detected from %d -> %d", ex.x, ex.y)
}

type ErrOutOfRange struct {
	s uint64
}

func (ex ErrOutOfRange) Error() string {
	return fmt.Sprintf("Value %d is out of range", ex.s)
}

type ErrOverlap struct {
	a, b uint64
}

func (ex ErrOverlap) Error() string {
	return fmt.Sprintf("Overlap detected between %d -> %d", ex.a, ex.b)
}

type ErrEmptyInput struct{}

func (ex ErrEmptyInput) Error() string {
	return "Input list is empty"
}

func NewRangeStoreFromWeighted(items []Weighted) (*Node, error) {
	if len(items) < 1 {
		return nil, ErrEmptyInput{}
	}
	totalWeight := uint64(0)
	ranges := make([]Ranged, 0)
	for _, item := range items {
		w := item.GetWeight()
		ranges = append(ranges, DefaultRangedValue{totalWeight + 1, totalWeight + w, item.GetValue()})
		newSum := totalWeight + w
		if newSum < totalWeight || newSum < w {
			return nil, ErrUnsignedIntegerOverflow{totalWeight, w}
		}
		totalWeight = newSum
	}

	return NewRangeStoreFromSorted(ranges)
}

// Builds a optimal(ish) tree containing the range values as
// node values. For the computation of optimality, we assume
// that every value in the aggregate ranges is equally likely
// to be looked up. We then choose pivots so that roughly equal
// amounts of range are in each subtree.
//
// That is, with a set of values like:
// * A [0,1]
// * B [1,2]
// * C [2,3]
// The tree produced will look like
//
//     B
//    [2]
//  ___|___
//  |     |
//  A     C
// [1]   [3]
//
// However with non-equal weights, such as
// * A [0,1]
// * B [1,2]
// * C [2,100]
// The tree produced will look more like
//
//        C
//      [100]
//     ___|
//     |
//     A
//    [1]
//     |___
//        |
//        B
//       [2]
//
// Although this tree is degenerate based on a counting of *nodes*
// it is optimal based on lookup frequency, since ~98% of lookups
// will terminate at the root node and only very infrequently will
// recursion down the tree occur.
//
// _Note_: The items passed to RangeStoreFRomSorted must contain
// a monotonically increasing and continuous sequence of min and
// max values
//
// _Note_: Construction of the tree is done using Mehlhorn's
// approximation for balancing the tree and uses an effective floor
// (unsigned integer division) when computing pivots. As a result,
// the produced data structure approaches, but may not always be
// exactly, optimal.
func NewRangeStoreFromSorted(items []Ranged) (*Node, error) {
	return rangeStoreFromSortedChecked(items, true)
}

// Helper function which takes a bool whether the ranges have already been checked
// Breaking this into two functions is a small optimization on big sets, but not
// having to do all of the overlap and discontinuity checking on every
// recursive call. We know that if we're calling recursively that we have
// only part of a range that's previously been through this function, so
// we can skip the checks for monotonicity.
func rangeStoreFromSortedChecked(items []Ranged, check bool) (*Node, error) {
	if len(items) < 1 {
		return nil, ErrEmptyInput{}
	}
	n := &Node{}
	// Easy base case: We've got one item. Just set it and forget it
	if len(items) == 1 {
		n.max = items[0].GetMax()
		n.value = items[0].GetValue()
	} else {
		// Compute the total weight in this slice
		// Also, check for discontinuities
		start := uint64(0)
		total := uint64(0)
		for idx, item := range items {
			if idx == 0 {
				start = item.GetMin()
			} else if check {
				// Check for discontinuity
				prev := items[idx-1].GetMax()
				curr := item.GetMin()
				if curr > prev+1 {
					return nil, ErrDiscontinuity{prev, curr}
				}
				// Check for overlap
				if curr < prev+1 {
					return nil, ErrOverlap{prev, curr}
				}
			}
			a := (item.GetMax() - item.GetMin()) + 1
			newSum := total + a
			if newSum < total || newSum < a {
				return nil, ErrUnsignedIntegerOverflow{total, a}
			}
			total = newSum
		}

		// Compute the pivot
		pivot := total / 2

		// Walk the list backwards and find the index of the item which has
		// a min less than the pivot
		var ridx int
		for ridx = len(items) - 1; ridx >= 0; ridx -= 1 {
			if items[ridx].GetMin() < pivot+start {
				break
			}
		}

		// Fill the node based on the current item
		n.max = items[ridx].GetMax()
		n.value = items[ridx].GetValue()

		// If we didn't pick the first item for the pivot, build the left subtree
		if ridx != 0 {
			// Explicitly ignore the error, since we've indicated we've already checked
			lft, _ := rangeStoreFromSortedChecked(items[:ridx], false)
			n.left = lft
		}
		// If we didn't pick the last item for the pivot, build the right subtree
		if ridx != len(items)-1 {
			// Explicitly ignore the error, since we've indicated we've already checked
			rht, _ := rangeStoreFromSortedChecked(items[ridx+1:], false)
			n.right = rht
		}
	}
	return n, nil
}

// Searches for the range which contains the specified key
// and returns the associated value, or an error if the
// value is out of range
func (n *Node) RangeSearch(val uint64) (interface{}, error) {
	if n.max < val {
		if n.right == nil {
			return nil, ErrOutOfRange{val}
		}
		return n.right.RangeSearch(val)
	} else {
		if n.left != nil {
			val, err := n.left.RangeSearch(val)
			if err == nil {
				return val, nil
			}
		}
		return n.value, nil
	}
}

// Creates a nicely formatter string representation of the Range Store. Useful for understanding how the data is
// internally stored and represented.
func (n *Node) String() string {
	return n.formattedString("")
}
func (n *Node) formattedString(prefix string) string {
	ret := fmt.Sprintf("%s-%s [max: %d]\n", prefix, n.value, n.max)
	if n.left != nil {
		ret += n.left.formattedString(prefix + " |")
	}
	if n.right != nil {
		ret += n.right.formattedString(prefix + " !")
	}
	return ret
}
