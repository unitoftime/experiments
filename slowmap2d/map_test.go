package main

import (
	"fmt"
	"testing"

	"github.com/unitoftime/ecs"
)

// This was comparing generic map functions vs regular map functions

type Vec2 struct {
	X, Y float64
}
type Position Vec2
type Collider struct {
	Radius float64
	Count int32
}

func BenchmarkMap2D(b *testing.B) {
	fmt.Println("Starting")
	fmt.Println("Starting")
	s := int(1000)
	a := &Archetypes[Position, Collider]{
		ids: [][]ecs.Id{make([]ecs.Id, s)},
		a: [][]Position{make([]Position, s)},
		b: [][]Collider{make([]Collider, s)},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		a.Map2D(func(
			aId []ecs.Id, aPos []Position, aCol []Collider,
			bId []ecs.Id, bPos []Position, bCol []Collider) {
				if len(aId) != len(aPos) || len(aId) != len(aCol) { panic("ERR") }
				if len(bId) != len(bPos) || len(bId) != len(bCol) { panic("ERR") }

				for i := range aId {
					for j := range bId {
						if aId[i] == bId[j] { continue } // Skip if entity is the same

						dx := aPos[i].X - bPos[j].X
						dy := aPos[i].Y - bPos[j].Y
						distSq := (dx * dx) + (dy * dy)

						dr := aCol[i].Radius + bCol[j].Radius
						drSq := dr * dr

						if drSq > distSq {
							aCol[i].Count++
						}
					}
				}
			})
	}
}

type Archetypes[A, B any] struct {
	ids [][]ecs.Id
	a [][]A
	b [][]B
}

func (a *Archetypes[A, B]) Map2D(f func([]ecs.Id, []A, []B, []ecs.Id, []A, []B)) {
	for i := range a.ids {
		for j := range a.ids {
			f(a.ids[i], a.a[i], a.b[i], a.ids[j], a.a[j], a.b[j])
		}
	}
}
