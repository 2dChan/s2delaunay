package s2voronoi

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

// Cell

func TestCell_SiteIndex(t *testing.T) {
	vd := mustNewDiagram(t, 100)
	for i := range vd.Sites {
		c := vd.Cell(i)
		if got := c.SiteIndex(); got != i {
			t.Errorf("c.SiteIndex() = %v, want %v", got, i)
		}
	}
}

func TestCell_Site(t *testing.T) {
	vd := mustNewDiagram(t, 100)
	for i, want := range vd.Sites {
		c := vd.Cell(i)
		if got := c.Site(); got != want {
			t.Errorf("c.SiteIndex() = %v, want %v", got, want)
		}
	}
}

func TestCell_NumVertices(t *testing.T) {
	vd := mustNewDiagram(t, 100)
	for i := range vd.Sites {
		c := vd.Cell(i)
		want := vd.CellOffsets[i+1] - vd.CellOffsets[i]
		if got := c.NumVertices(); got != want {
			t.Errorf("c.NumVertices() = %v, want %v", got, want)
		}
	}
}

func TestCell_VertexIndices(t *testing.T) {
	vd := mustNewDiagram(t, 100)
	for i := range vd.Sites {
		c := vd.Cell(i)
		want := vd.CellVertices[vd.CellOffsets[i]:vd.CellOffsets[i+1]]
		got := c.VertexIndices()
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("c.VertexIndices() mismatch (-want +got):\n%v", diff)
		}
	}
}

func TestCell_Vertex(t *testing.T) {
	vd := mustNewDiagram(t, 100)
	for i := range vd.Sites {
		c := vd.Cell(i)
		indices := c.VertexIndices()
		for j, idx := range indices {
			want := vd.Vertices[idx]
			got := c.Vertex(j)
			if got != want {
				t.Errorf("c.Vertex(%d) = %v, want %v", j, got, want)
			}
		}
	}
}

func TestCell_Vertex_Panic(t *testing.T) {
	d := mustNewDiagram(t, 100)
	c := d.Cell(0)
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("c.VertexIndices should panic for i out of range")
		}
	}()
	c.Vertex(-1)
	c.Vertex(len(d.CellOffsets) + 1)
}

func TestCell_NumNeighbors(t *testing.T) {
	vd := mustNewDiagram(t, 100)
	for i := range vd.Sites {
		c := vd.Cell(i)
		want := vd.CellOffsets[i+1] - vd.CellOffsets[i]
		if got := c.NumNeighbors(); got != want {
			t.Errorf("c.NumNeighbors() = %v, want %v", got, want)
		}
	}
}

func TestCell_NeighborIndices(t *testing.T) {
	vd := mustNewDiagram(t, 100)
	for i := range vd.Sites {
		c := vd.Cell(i)
		want := vd.CellNeighbors[vd.CellOffsets[i]:vd.CellOffsets[i+1]]
		got := c.NeighborIndices()
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("c.NeighborIndices() mismatch (-want +got, cell %d):\n%v", i, diff)
		}
	}
}

func TestCell_Neighbor(t *testing.T) {
	vd := mustNewDiagram(t, 100)
	for i := range vd.Sites {
		c := vd.Cell(i)
		neighbors := c.NeighborIndices()
		for j, nIdx := range neighbors {
			got := c.Neighbor(j)
			if got.SiteIndex() != nIdx {
				t.Errorf("c.Neighbor(%d).SiteIndex() = %v, want %v", j, got.SiteIndex(), nIdx)
			}
		}
	}
}

func TestCell_Neighbor_Panic(t *testing.T) {
	d := mustNewDiagram(t, 100)
	d.CellOffsets = []int{0, 1}
	c := d.Cell(0)
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for invalid Neighbor indices, but did not panic")
		}
	}()
	c.Neighbor(-1)
	c.Neighbor(len(d.CellOffsets) + 1)
}
