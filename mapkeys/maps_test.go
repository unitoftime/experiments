package mapkeys

import (
	// "fmt"
	"testing"
)

// Compare if it's faster to use primitives for map keys or structs for map keys

type PrimKey uint64
type StructKey struct {
	A, B int
}

type ValueStruct struct {
	Data [10]int
}

func BenchmarkPrimitiveKeyWrite(b *testing.B) {
	m := make(map[PrimKey]ValueStruct)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m[PrimKey(i)] = ValueStruct{}
	}
}

func BenchmarkStructKeyWrite(b *testing.B) {
	m := make(map[StructKey]ValueStruct)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m[StructKey{i,0}] = ValueStruct{}
	}
}
