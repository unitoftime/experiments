package iterators

import (
	"testing"
)
// I started this investigation because I was seeing some weird stuff with looping. Specifically
// I was working on my ecs library and was trying to come up with a faster iteration method so
// that I could compare my ECS performance to that of Bevy's (A Rust ECS library).
// During that investigation I saw some weird thing. This file is about those weird things, and
// also it's about my seemingly neverending attempt to build a fast, generic iterator for my ecs library

// Let's introduce a few things we will use for our benchmark
// An ECS Id perhaps?
type Id uint64

// A 3d position
type Position struct {
	X, Y, Z float32
}

// A 3d velocity
type Velocity struct {
	X, Y, Z float32
}

// Just a fake multiplier for time based-movement
var dt = float32(0.001)

// This is the physics function we're executing: move the position by the velocity times the delta time change. From the perspective of a developer, they'd write this function and they'd want the underlying system to execute as quickly as possible; essentially mapping their function across their dataset (presumably held in the ECS).
func physicsTick(id Id, pos *Position, vel *Velocity) {
	pos.X += vel.X * dt
	pos.Y += vel.Y * dt
	pos.Z += vel.Z * dt
}

// Here we define a map function (in the functional-programming sense) that operates on pre-specified types: Id, Position, and Velocity.
func mapFuncPhy(id []Id, pos []Position, vel []Velocity, f func(id Id, pos *Position, vel *Velocity)) {
	for j := range id {
		f(id[j], &pos[j], &vel[j])
	}
}

// Our objective will be to make this generic version as fast as possible (essentially comparing it to the statically generated map function `mapFuncPhy`
func mapFuncPhyGen[A any, B any](id []Id, aa []A, bb []B, f func(id Id, x *A, y *B)) {
	for j := range id {
		f(id[j], &aa[j], &bb[j])
	}
}


// I'm going to pack all of the data into a struct, just so that we have some sort of dataset that we are operating on
type Data struct {
	ids []Id
	pos []Position
	vel []Velocity
}

// Let's give everything 1 million entities
func NewData() *Data {
	return &Data{
		ids: make([]Id, 1e6),
		pos: make([]Position, 1e6),
		vel: make([]Velocity, 1e6),
	}
}

// Now we can start benchmarking

// First things first, we'll make this as efficient as possible. Suppose we literally just loop over the data in our non-generic mapFuncPhy mapper function.
func BenchmarkData(b *testing.B) {
	d := NewData()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		mapFuncPhy(d.ids, d.pos, d.vel, physicsTick)
	}
}

// Actually, that might not be the fastest - Because we are passing in the physicsTick function into an extra function, so maybe that makes it slower? Let's also create this even-more-simplified benchmark to compare
func BenchmarkDataShouldBeFastest(b *testing.B) {
	d := NewData()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for j := range d.ids {
			physicsTick(d.ids[j], &d.pos[j], &d.vel[j])
		}
	}
}

// Very surprisingly, the results I see on my computer are contrary to what I originally assumed.
// cpu: Intel(R) Core(TM) i7-8700K CPU @ 3.70GHz
// BenchmarkData-12                           	     772	   1515732 ns/op
// BenchmarkDataShouldBeFastest-12            	     578	   2030759 ns/op

// After looking and comparing the assembly of these two benchmarks, the only notable difference that I could found was that in `BenchmarkDataShouldBeFastest` was checking slice bounds more often than `BenchmarkData`: `BenchmarkData` checks 2 bounds, `BenchmarkDataShouldBeFastest` is checking 3 bounds. I'm not yet sure why this is happening in this benchmark. So let's keep looking. I was able to learn a bit about bounds checking from this [webpage](https://go101.org/article/bounds-check-elimination.html). Also, there's some interesting reading [here](https://docs.google.com/document/d/1vdAEAjYdzjnPA9WDOQ1e4e05cYVMpqSxJYZT33Cqw2g/edit).
// I think this implies something else: which is that in the `BenchmarkDataShouldBeFastest` benchmark, we have to index into d.ids[j] and pass that id into the function. I think the compiler isn't optimizing this out for some reason, even though it's technically dead code.

// Also notably, we can run this to see which bounds checks are in place:
// `go test . -gcflags="-d=ssa/check_bce/debug=1" -bench=Data`

// Also notably, we can remove bounds checks from code: `go test -gcflags=-B . -bench=Data`
// BenchmarkData-12                                       	     786	   1492962 ns/op
// BenchmarkDataShouldBeFastest-12                        	     681	   1733128 ns/op
// BenchmarkDataShouldBeFastestNoStruct-12                	     788	   1497137 ns/op

// We can build another benchmark that doesn't use the struct, just to see how she does.
func BenchmarkDataShouldBeFastestNoStruct(b *testing.B) {
	ids := make([]Id, 1e6)
	pos := make([]Position, 1e6)
	vel := make([]Velocity, 1e6)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for j := range ids {
			physicsTick(ids[j], &pos[j], &vel[j])
		}
	}
}

// This benchmark ends up doing just as fast as the first one:
// cpu: Intel(R) Core(TM) i7-8700K CPU @ 3.70GHz
// BenchmarkData-12                           	     771	   1504308 ns/op
// BenchmarkDataShouldBeFastest-12            	     571	   2026442 ns/op
// BenchmarkDataShouldBeFastestNoStruct-12    	     746	   1509959 ns/op

// I'm not really sure how not putting the slices into structs will affect the bounds check elimination that we want to have. Let's add another benchmark to see.
func BenchmarkDataShouldBeFastestHybridStructNoStruct(b *testing.B) {
	d := NewData()
	ids := d.ids
	pos := d.pos
	vel := d.vel

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for j := range ids {
			physicsTick(ids[j], &pos[j], &vel[j])
		}
	}
}

// Here is the current tally:
// cpu: Intel(R) Core(TM) i7-8700K CPU @ 3.70GHz
// BenchmarkData-12                                       	     780	   1508880 ns/op
// BenchmarkDataShouldBeFastest-12                        	     585	   2013123 ns/op
// BenchmarkDataShouldBeFastestNoStruct-12                	     764	   1518559 ns/op
// BenchmarkDataShouldBeFastestHybridStructNoStruct-12    	     784	   1508425 ns/op

// One second, I'm going to peek at the assembly files again now to see if all of these actually are able to do the bounds check elimination. We can do that like this [link](https://go.dev/doc/gdb):
// Generate unit test binary file: `go test -c`
// Dump the ASM to a file: `go tool objdump -S iterators.test > iterators.obj`
// Read (Caution: It's big)

// After some reading, it's really hard to tell where bounds checks happen. But at least some of the performance is reduced because of bounds checking. Let's generate the assembly with the bounds checks disabled just to see what it looks like:
// go test  -gcflags="-B" -c

// Notably the benchmarks look like this now (after removing branch checking):
// BenchmarkData-12                                       	     777	   1485453 ns/op
// BenchmarkDataShouldBeFastest-12                        	     673	   1724847 ns/op
// BenchmarkDataShouldBeFastestNoStruct-12                	     787	   1496575 ns/op
// BenchmarkDataShouldBeFastestHybridStructNoStruct-12    	     757	   1506834 ns/op
// So there is still something that slows down `BenchmarkDataShouldBeFastest`

// After generating it without benchmarks, the differences are even more minor. Specifically, there is some difference like this:
// < // Left side this is BenchmarkData
// <  0x4ef890		488b742440		MOVQ 0x40(SP), SI
// <  0x4ef895		488b7c2458		MOVQ 0x58(SP), DI
// < for j := range ids {

// < // Right side this is BenchmarkDataShouldBeFastest
// > for j := range ids {
// >  0x4ef6e2		488b5c2440		MOVQ 0x40(SP), BX
// >  0x4ef6e7		488b742458		MOVQ 0x58(SP), SI

// I'm *pretty* sure that this is indicating that the slice variables aren't automatically hoisted to the top of the for loop *but only* when the slices are queried from the location in the struct. To be honest, I'm not really sure why this is - but mystery solved, at least I think.

// Lesson Learned: Hoist all of your slice variables manually?
