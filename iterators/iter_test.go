// Package iterators is a testbed for various iterator-like implementations
// Save this to a file named iter_test.go and create a dummy file iter.go with just the line
//
//   // file iter.go
//   package iterators
//
// Then run:
//
//   go test .
//   go test -bench .
//
// Results:
//	BenchmarkArrayLike       500	   2492641 ns/op
//	BenchmarkClosure         200	   7966233 ns/op
//	BenchmarkClosure2        100	  10322898 ns/op
//	BenchmarkClosure3        200	   8512510 ns/op
//	BenchmarkStateful        300	   4010607 ns/op
//	BenchmarkStatefulErr     300	   4887407 ns/op

package iterators_test

import (
	"errors"
	"fmt"
	"testing"
)

// errors.New takes an awful lot of cycles. Better setup errors outside of time critical loops
var errDone = errors.New("Done")

// Collection is a dummy collection.
// For the sake of testing, we'll store items in reverse order
type Collection struct {
	items []int
}

// Push adds an item at the beginning of the collection
func (c *Collection) Push(n int) {
	// Items stored in reverse order, so append(.., n) makes it the first item
	c.items = append(c.items, n)
}

// NewCollection creates a new collection initialized with numbers from 1 to n.
func NewCollection(n int) *Collection {
	var c Collection
	for i := n; i > 0; i-- {
		c.Push(i)
	}
	return &c
}

// ------ Raw for loop ------
func BenchmarkArrayLike(b *testing.B) {
	c := NewCollection(1e6)
	b.ResetTimer()
	l := c.Len()
	for i := 0; i < b.N; i++ {
		for n := 0; n < l; n++ {
			c.items[n] = c.items[n] + 1 // Do something
		}
	}
}

// ------ Array-style accessors ------

// ValueAt returns the Nth item.
func (c *Collection) ValueAt(n int) int {
	return c.items[len(c.items)-n-1] // let it panic if out of bounds
}

func (c *Collection) SetValue(n int) int {
	c.items[len(c.items)-n-1] // let it panic if out of bounds
}

// Len returns the number of items in the collection
func (c *Collection) Len() int { return len(c.items) }

func BenchmarkArrayLike(b *testing.B) {
	c := NewCollection(1e6)
	b.ResetTimer()
	l := c.Len()
	for i := 0; i < b.N; i++ {
		for n := 0; n < l; n++ {
			c.SetValue(n, c.ValueAt(n) + 1)
		}
	}
}

// ------ Closure iterators ------

// ClosureIter returns a closure based iterator and a boolean set to true if there
// are values to be read.
func (c *Collection) ClosureIter() (f func() (int, bool), hasNext bool) {
	l := len(c.items)
	hasNext = l > 0
	f = func() (int, bool) {
		l--
		return c.items[l], l > 0
	}
	return
}

// ClosureIter2 is almost the same as above but returns a next() and hasNext() function (just to
// show that this approach is slower).
func (c *Collection) ClosureIter2() (next func() int, hasNext func() bool) {
	l := len(c.items)
	next = func() int { l--; return c.items[l] }
	hasNext = func() bool { return l > 0 }
	return
}

// ClosureIter3 is a variation on ClosureIter that returns an error when trying to read past
// the end of the collection. It's slower than ClosureIter that uses a "predictive" hasNext, thus
// skipping a conditional branch.
func (c *Collection) ClosureIter3() (next func() (int, error)) {
	l := len(c.items)
	return func() (int, error) {
		if l > 0 {
			l--
			return c.items[l], nil
		}
		return 0, errDone
	}
}

// Tests
func BenchmarkClosure(b *testing.B) {
	c := NewCollection(1e6)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for next, hasNext := c.ClosureIter(); hasNext; {
			_, hasNext = next()
		}
	}
}

func BenchmarkClosure2(b *testing.B) {
	c := NewCollection(1e6)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for next, hasNext := c.ClosureIter2(); hasNext(); {
			_ = next()
		}
	}
}

func BenchmarkClosure3(b *testing.B) {
	c := NewCollection(1e6)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter := c.ClosureIter3()
		for _, err := iter(); err == nil; _, err = iter() {
			// do something with v?
		}
	}
}

// ------ Stateful iterator ------

type iterator2 struct {
	*Collection // neat!
	index       int
}

func (i *iterator2) Next(val *int) bool {
	i.index--
	*val = i.items[i.index]
	return i.index > 0
}

// StatefulIter returns our stateful iterator.
func (c *Collection) StatefulIter2() *iterator2 {
	return &iterator2{c, len(c.items)}
}

func BenchmarkStateful2(b *testing.B) {
	c := NewCollection(1e6)
	b.ResetTimer()
	var val int
	for i := 0; i < b.N; i++ {
		// for iter := c.StatefulIter2(); iter.HasNext(); {
		iter := c.StatefulIter2();
		for iter.Next(&val) {
		}
	}
}




// this iterator struct does not have to be exported, but can be if you have some
// generic Iterator interface that you want/need to implement.
type iterator struct {
	*Collection // neat!
	index       int
}

// Next returns the next item in the collection.
func (i *iterator) Next() int {
	i.index--
	return i.items[i.index]
}

// NextErr is a variation of HasNext that returns an error when trying to read past
// the end of the collection (no need to use HasNext). It is slower with this particular
// Collection implementation since it introduces an additional conditional branch.
// More complex collections may not suffer that much.
func (i *iterator) NextErr() (int, error) {
	if i.index == 0 {
		return 0, errDone
	}
	i.index--
	return i.items[i.index], nil
}

// HasNext return true if there are values to be read.
func (i *iterator) HasNext() bool {
	return i.index > 0
}

// StatefulIter returns our stateful iterator.
func (c *Collection) StatefulIter() *iterator {
	return &iterator{c, len(c.items)}
}

// Tests

func ExampleStateful() {
	c := NewCollection(5)
	for iter := c.StatefulIter(); iter.HasNext(); {
		fmt.Printf("%d ", iter.Next())
	}
	// Output:
	// 1 2 3 4 5
}

func ExampleStatefulErr() {
	c := NewCollection(5)
	iter := c.StatefulIter()
	for v, err := iter.NextErr(); err == nil; v, err = iter.NextErr() {
		fmt.Printf("%d ", v)
	}
	// Output:
	// 1 2 3 4 5
}

func BenchmarkStateful(b *testing.B) {
	c := NewCollection(1e6)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for iter := c.StatefulIter(); iter.HasNext(); {
			_ = iter.Next()
		}
	}
}

func BenchmarkStatefulErr(b *testing.B) {
	c := NewCollection(1e6)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter := c.StatefulIter()
		for v, err := iter.NextErr(); err == nil; v, err = iter.NextErr() {
			_ = v
		}
	}
}
