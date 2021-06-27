package alignment

import (
	"github.com/rschio/align/alignment/internal/bitbucket"
	"github.com/rschio/graph"
)

type Graph struct {
	Interface
	Src, Dst  int
	Edges     [][]int
	Labels    []rune
	SeqLabels []rune
	Loops     []bool
	Score     ScoreFn
	order     int
}

// Assert, in compile time, Graph satisfies
// the graph.Iterator interface.
var _ graph.Iterator = (*Graph)(nil)

type Interface interface {
	VisitFromSrc(do func(w int, c int64) bool) bool
	VisitFromLastRow(v int, do func(w int, c int64) bool) bool
}

func (g *Graph) ShortestPath() (path []int, dist int64) {
	q := new(bitbucket.Queue)
	return graph.ShortestPathWithQueue(g, q, g.Src, g.Dst)
}

func (g *Graph) Order() int {
	return g.order
}

func (g *Graph) Visit(v int, do func(w int, c int64) bool) bool {
	if !g.normalRow(v) {
		if v == g.Dst {
			return false
		}
		if v == g.Src {
			return g.VisitFromSrc(do)
		}
		return g.VisitFromLastRow(v, do)
	}
	vertices := len(g.Labels)
	vi := v % vertices
	row := v / vertices
	offset := v - vi
	vertical := vi + vertices + offset
	// Process horizontal edges.
	var i, w int
	for i, w = range g.Edges[vi] {
		w += offset
		// Use vertical edge as sentinel to
		// split the edges in
		// horizontals:vertical:diagonals.
		if w == vertical {
			break
		}
		// g.Score('A', space) always have a
		// positive cost.
		c := g.Score('A', space)
		if do(w, c) {
			return true
		}
	}
	// Process the vertical one.
	//
	// This is a tricky part.
	// If this vertex has no loop, get its cost that
	// will be always > 0 and advance to next edge.
	// But if this vertex does have a loop, do not
	// process it here, treat the edge as a diagonal
	// edge.
	if g.Loops[vi] == false {
		c := g.Score('A', space)
		if do(w, c) {
			return true
		}
		i++
	}
	// When the vertex has no horizontal
	// neighbor and the only edge that exists
	// is the vertical one already treat above.
	if i >= len(g.Edges[vi]) {
		return false
	}
	// Process the diagonal edges.
	diagonals := g.Edges[vi][i:]
	for _, w = range diagonals {
		w += offset
		p := w % vertices
		c := g.Score(g.SeqLabels[row+1], g.Labels[p])
		if do(w, c) {
			return true
		}
	}
	return false
}

func (g *Graph) normalRow(v int) bool {
	vertices := len(g.Labels)
	row := v / vertices
	return row < len(g.SeqLabels)-1
}
