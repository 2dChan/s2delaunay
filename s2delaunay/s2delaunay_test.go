// Copyright (c) 2026 Andrey Kriulin
// Licensed under the MIT License.
// See the LICENSE file in the project root for full license text.

package s2delaunay

import (
	"fmt"
	"math"
	"testing"

	"github.com/2dChan/s2voronoi/utils"
	"github.com/golang/geo/r3"
	"github.com/golang/geo/s2"
	"github.com/markus-wa/quickhull-go/v2"
)

// Triangle

func TestTrianglePrevVertex(t *testing.T) {
	verts := [3]int{1, 2, 3}
	tri := Triangle{V: verts}
	for i, v := range tri.V {
		got := tri.PrevVertex(v)
		want := verts[(i+2)%len(tri.V)]
		if got != want {
			t.Errorf("PrevVertex: got %d, want %d", got, want)
		}
	}

	got1 := tri.PrevVertex(-1)
	want1 := -1
	if got1 != want1 {
		t.Errorf("PrevVertex: got %d, want %d", got1, want1)
	}
}

func TestTriangleNextVertex(t *testing.T) {
	verts := [3]int{1, 2, 3}
	tri := Triangle{V: verts}
	for i, v := range tri.V {
		got := tri.NextVertex(v)
		want := verts[(i+1)%len(tri.V)]
		if got != want {
			t.Errorf("NextVertex: got %d, want %d", got, want)
		}
	}

	got1 := tri.NextVertex(-1)
	want1 := -1
	if got1 != want1 {
		t.Errorf("NextVertex: got %d, want %d", got1, want1)
	}
}

// DelaunayTriangulation

func BenchmarkConvexHull(b *testing.B) {
	sizes := []int{1e+2, 1e+3, 1e+4, 1e+5}
	for _, pointsCnt := range sizes {
		b.Run(fmt.Sprintf("N%d", pointsCnt), func(b *testing.B) {
			points := utils.GenerateRandomPoints(pointsCnt, 0)
			vertices := pointsToVectors(points)
			qh := new(quickhull.QuickHull)

			b.ResetTimer()
			for b.Loop() {
				qh.ConvexHull(vertices, true, true, 0)
			}
		})
	}
}

func pointsToVectors(points []s2.Point) []r3.Vector {
	vectors := make([]r3.Vector, len(points))
	for i, p := range points {
		vectors[i] = p.Vector
	}
	return vectors
}

func BenchmarkComputeDelaunayTriangulation(b *testing.B) {
	sizes := []int{1e+2, 1e+3, 1e+4, 1e+5}
	for _, pointsCnt := range sizes {
		b.Run(fmt.Sprintf("N%d", pointsCnt), func(b *testing.B) {
			points := utils.GenerateRandomPoints(pointsCnt, 0)

			b.ResetTimer()
			for b.Loop() {
				_, err := ComputeDelaunayTriangulation(points, 0)
				if err != nil {
					b.Fatalf("ComputeDelaunayTriangulation: got error %v, want nil.", err)
				}
			}
		})
	}
}

func TestIncidentTriangles_OutOfRange(t *testing.T) {
	dt := &DelaunayTriangulation{
		Vertices:                nil,
		Triangles:               nil,
		IncidentTriangleIndices: nil,
		IncidentTriangleOffsets: []int{0, 0, 0},
	}

	if _, err := dt.IncidentTriangles(-1); err == nil {
		t.Error("IncidentTriangles: got nil, want error for vIdx = -1")
	}

	if _, err := dt.IncidentTriangles(2); err == nil {
		t.Error("IncidentTriangles: got nil, want error for vIdx = 2")
	}
}

func TestTriangleVertices(t *testing.T) {
	points := utils.GenerateRandomPoints(3, 0)
	dt := &DelaunayTriangulation{
		Vertices: s2.PointVector{points[0], points[1], points[2]},
		Triangles: []Triangle{
			{V: [3]int{0, 1, 2}},
		},
	}
	want := [3]s2.Point{points[0], points[1], points[2]}
	got, err := dt.TriangleVertices(0)
	if err != nil || got != want {
		t.Errorf("TriangleVertices(0): got %v, %v; want %v, nil", got, err, want)
	}
	_, err = dt.TriangleVertices(1)
	if err == nil {
		t.Error("TriangleVertices(1): got nil, want error")
	}
	_, err = dt.TriangleVertices(-1)
	if err == nil {
		t.Error("TriangleVertices(-1): got nil, want error")
	}
}

func TestComputeDelaunayTriangulation_DegenerateInput(t *testing.T) {
	vertices := s2.PointVector{
		s2.PointFromCoords(1, 0, 0),
		s2.PointFromCoords(0, 1, 0),
		s2.PointFromCoords(0, 0, 1),
	}
	_, err := ComputeDelaunayTriangulation(vertices, 0)
	if err == nil {
		t.Errorf("ComputeDelaunayTriangulation (degenerate input): got nil, want error")
	}
}

func TestComputeDelaunayTriangulation_VerticesOnSphere(t *testing.T) {
	const (
		eps = defaultEps
	)

	vertices := utils.GenerateRandomPoints(100, 0)
	dt, err := ComputeDelaunayTriangulation(vertices, eps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i, p := range dt.Vertices {
		norm := p.Norm()
		if math.Abs(norm-1.0) > eps {
			t.Errorf("ComputeDelaunayTriangulation: point %d norm = %f, expected ~1.0 (on sphere)",
				i, norm)
		}
	}
}

func TestComputeDelaunayTriangulation_VerifyIncidentTrianglesSorted(t *testing.T) {
	vertices := utils.GenerateRandomPoints(100, 0)
	dt, err := ComputeDelaunayTriangulation(vertices, 0)
	if err != nil {
		t.Error(err)
	}

	for vIdx := range len(dt.Vertices) {
		incidentTris, _ := dt.IncidentTriangles(vIdx)
		for i := 1; i < len(incidentTris); i++ {
			currentTri := dt.Triangles[incidentTris[i-1]]
			nextTri := dt.Triangles[incidentTris[i]]

			nextVertex := currentTri.NextVertex(vIdx)
			prevVertex := nextTri.PrevVertex(vIdx)

			if nextVertex != prevVertex {
				t.Errorf("Incident triangles not sorted CCW for vertex %d: tris %d,%d",
					vIdx, incidentTris[i-1], incidentTris[i])
			}
		}
	}
}

func TestSortTriangleVerticesCCW(t *testing.T) {
	a := s2.PointFromCoords(1, 0, 0)
	b := s2.PointFromCoords(0, 1, 0)
	c := s2.PointFromCoords(0, 0, 1)
	verts := s2.PointVector{a, b, c}

	tri1 := &Triangle{V: [3]int{0, 1, 2}}
	sortTriangleVerticesCCW(tri1, verts)
	// Already CCW, should not change
	if tri1.V != [3]int{0, 1, 2} {
		t.Errorf("sortTriangleVerticesCCW: got %v, want {0,1,2}", tri1.V)
	}

	tri2 := &Triangle{V: [3]int{0, 2, 1}}
	sortTriangleVerticesCCW(tri2, verts)
	// Should reorder to CCW
	if tri2.V != [3]int{0, 1, 2} {
		t.Errorf("sortTriangleVerticesCCW: got %v, want {0,1,2}", tri2.V)
	}
}

func TestSortIncidentTriangleIndicesCCW(t *testing.T) {
	expected3 := []int{0, 1, 2}
	incident3 := []int{1, 0, 2}
	tris3 := []Triangle{
		{V: [3]int{0, 1, 2}},
		{V: [3]int{0, 2, 3}},
		{V: [3]int{0, 3, 1}},
	}
	sortIncidentTriangleIndicesCCW(0, incident3, tris3)
	if cyclicEqual(incident3, expected3) {
		t.Errorf("sortIncidentTriangleIndicesCCW: got %v, want %v", incident3, expected3)
	}

	expected4 := []int{0, 1, 2, 3}
	incident4 := []int{3, 2, 1, 0}
	tris4 := []Triangle{
		{V: [3]int{0, 1, 2}},
		{V: [3]int{0, 2, 3}},
		{V: [3]int{0, 3, 4}},
		{V: [3]int{0, 4, 1}},
	}
	sortIncidentTriangleIndicesCCW(0, incident4, tris4)
	if cyclicEqual(incident4, expected4) {
		t.Errorf("sortIncidentTriangleIndicesCCW: got %v, want %v", incident4, expected4)
	}
}

func cyclicEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}

	n := len(a)
	for i := range n {
		if b[0] != a[i] {
			continue
		}

		equal := true
		for j := range n {
			if a[(i+j)%n] != b[j] {
				equal = false
				break
			}
		}
		if equal {
			return true
		}
	}

	return false
}
