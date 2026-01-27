// Copyright (c) 2026 Andrey Kriulin
// Licensed under the MIT License.
// See the LICENSE file in the project root for full license text.

package s2voronoi

import (
	"github.com/2dChan/s2voronoi/s2delaunay"
	"github.com/golang/geo/s2"
)

const (
	defaultEps = 1e-12
)

type Diagram struct {
	Sites    s2.PointVector
	Vertices s2.PointVector

	// NOTE: Sort in CCW per Cell(look out of sphere)
	CellVertices []int
	// NOTE: Sort in CCW per Cell(look out of sphere)
	CellNeighbors []int
	CellOffsets   []int
}

func NewDiagram(sites s2.PointVector, eps float64) (*Diagram, error) {
	if eps == 0 {
		eps = defaultEps
	}

	dt, err := s2delaunay.NewTriangulation(sites, s2delaunay.WithEps(eps))
	if err != nil {
		return nil, err
	}

	numTriangles := len(dt.Triangles)
	numNeighbors := len(dt.IncidentTriangleIndices)
	d := &Diagram{
		Sites:         dt.Vertices,
		Vertices:      make(s2.PointVector, numTriangles),
		CellVertices:  dt.IncidentTriangleIndices,
		CellNeighbors: make([]int, numNeighbors),
		CellOffsets:   dt.IncidentTriangleOffsets,
	}

	for i := range numTriangles {
		p0, p1, p2 := dt.TriangleVertices(i)
		d.Vertices[i] = s2.Point{Vector: triangleCircumcenter(p0, p1, p2).Normalize()}
	}

	for vIdx := range dt.Vertices {
		offset := dt.IncidentTriangleOffsets[vIdx]
		it := dt.IncidentTriangles(vIdx)
		for i, tIdx := range it {
			d.CellNeighbors[offset+i] = s2delaunay.NextVertex(dt.Triangles[tIdx], vIdx)
		}
	}

	return d, nil
}

func (d *Diagram) NumCells() int {
	return len(d.Sites)
}

func (d *Diagram) Cell(i int) Cell {
	return Cell{idx: i, d: d}
}

func triangleCircumcenter(p1, p2, p3 s2.Point) s2.Point {
	v1 := p1.Sub(p2.Vector)
	v2 := p2.Sub(p3.Vector)

	circumcenter := v1.Cross(v2)

	if circumcenter.Dot(p1.Vector.Add(p2.Vector).Add(p3.Vector)) < 0 {
		circumcenter = circumcenter.Mul(-1)
	}

	return s2.Point{Vector: circumcenter}
}
