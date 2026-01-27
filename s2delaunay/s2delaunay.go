// Copyright (c) 2026 Andrey Kriulin
// Licensed under the MIT License.
// See the LICENSE file in the project root for full license text.

package s2delaunay

import (
	"errors"
	"fmt"

	"github.com/golang/geo/r3"
	"github.com/golang/geo/s2"
	"github.com/markus-wa/quickhull-go/v2"
)

const (
	defaultEps = 1e-12
)

type Triangulation struct {
	Vertices                s2.PointVector
	Triangles               [][3]int
	IncidentTriangleIndices []int
	IncidentTriangleOffsets []int
}

type TriangulationOptions struct {
	Eps float64
}

type TriangulationOption func(*TriangulationOptions) error

func WithEps(eps float64) TriangulationOption {
	return func(o *TriangulationOptions) error {
		if eps <= 0 {
			return fmt.Errorf("WithEps: eps must be positive got %v", eps)
		}
		o.Eps = eps
		return nil
	}
}

func NewTriangulation(vertices s2.PointVector, setters ...TriangulationOption) (*Triangulation,
	error) {
	opts := TriangulationOptions{
		Eps: defaultEps,
	}
	for _, set := range setters {
		err := set(&opts)
		if err != nil {
			return nil, err
		}
	}
	numVertices := len(vertices)
	if numVertices < 4 {
		return nil,
			errors.New("NewTriangulation: insufficient vertices for triangulation minimum 4 required")
	}
	numTriangles := 2 * (numVertices - 2)
	t := &Triangulation{
		Vertices:                vertices,
		Triangles:               make([][3]int, numTriangles),
		IncidentTriangleIndices: make([]int, numTriangles*3),
		IncidentTriangleOffsets: make([]int, numVertices+1),
	}
	r3vertices := make([]r3.Vector, numVertices)
	for i, p := range vertices {
		r3vertices[i] = p.Vector
	}
	qh := new(quickhull.QuickHull)
	ch := qh.ConvexHull(r3vertices, true, true, opts.Eps)
	if len(ch.Indices) != numTriangles*3 {
		return nil,
			errors.New("NewTriangulation: inconsistent number of indices returned from QuickHull")
	}
	for _, idx := range ch.Indices {
		t.IncidentTriangleOffsets[idx+1]++
	}
	for i := range numVertices {
		t.IncidentTriangleOffsets[i+1] += t.IncidentTriangleOffsets[i]
	}
	nxt := make([]int, numVertices)
	copy(nxt, t.IncidentTriangleOffsets[:numVertices])
	for i := range numTriangles {
		base := i * 3
		for j := range 3 {
			v := ch.Indices[base+j]
			t.Triangles[i][j] = v
			t.IncidentTriangleIndices[nxt[v]] = i
			nxt[v]++
		}
		sortTriangleVerticesCCW(&t.Triangles[i], t.Vertices)
	}
	for i := range numVertices {
		incidentTriangles, err := t.IncidentTriangles(i)
		if err != nil {
			return nil, err
		}
		sortIncidentTriangleIndicesCCW(i, incidentTriangles, t.Triangles)
	}
	return t, nil
}

func (t *Triangulation) IncidentTriangles(vIdx int) ([]int, error) {
	if vIdx < 0 || vIdx+1 >= len(t.IncidentTriangleOffsets) {
		return nil,
			fmt.Errorf("IncidentTriangles: vIdx %d out of range [0 %d)", vIdx,
				len(t.IncidentTriangleOffsets)-1)
	}
	start := t.IncidentTriangleOffsets[vIdx]
	end := t.IncidentTriangleOffsets[vIdx+1]
	return t.IncidentTriangleIndices[start:end], nil
}

func (t *Triangulation) TriangleVertices(tIdx int) ([3]s2.Point, error) {
	if tIdx < 0 || tIdx >= len(t.Triangles) {
		return [3]s2.Point{},
			fmt.Errorf("TriangleVertices: tIdx %d out of bounds [0 %d)", tIdx, len(t.Triangles))
	}
	tri := t.Triangles[tIdx]
	return [3]s2.Point{t.Vertices[tri[0]], t.Vertices[tri[1]], t.Vertices[tri[2]]}, nil
}

func sortTriangleVerticesCCW(t *[3]int, v s2.PointVector) {
	p0, p1, p2 := v[t[0]], v[t[1]], v[t[2]]
	norm := p1.Sub(p0.Vector).Cross(p2.Sub(p0.Vector))
	if norm.Dot(p0.Vector) < 0 {
		t[1], t[2] = t[2], t[1]
	}
}

func sortIncidentTriangleIndicesCCW(vIdx int, incidentTris []int, tris [][3]int) {
	n := len(incidentTris)
	for i := 1; i < n; i++ {
		nxt, err := NextVertex(tris[incidentTris[i-1]], vIdx)
		if err != nil {
			panic(err)
		}
		for j := i + 1; j < n; j++ {
			prv, err := PrevVertex(tris[incidentTris[j]], vIdx)
			if err != nil {
				panic(err)
			}
			if nxt == prv {
				incidentTris[i], incidentTris[j] = incidentTris[j], incidentTris[i]
				break
			}
		}
	}
}

func PrevVertex(t [3]int, vIdx int) (int, error) {
	switch vIdx {
	case t[0]:
		return t[2], nil
	case t[1]:
		return t[0], nil
	case t[2]:
		return t[1], nil
	}
	return 0, fmt.Errorf("PrevVertex: vIdx %d not in triangle", vIdx)
}

func NextVertex(t [3]int, vIdx int) (int, error) {
	switch vIdx {
	case t[0]:
		return t[1], nil
	case t[1]:
		return t[2], nil
	case t[2]:
		return t[0], nil
	}
	return 0, fmt.Errorf("NextVertex: vIdx %d not in triangle", vIdx)
}
