package alignment

import (
	"github.com/rschio/align/parse"
)

const space = '-'

type ScoreFn func(a, b rune) int64

type Base struct {
	Src, Dst  int
	Edges     [][]int
	Labels    []rune
	SeqLabels []rune
	Loops     []bool
	Score     ScoreFn
	order     int
}

func NewBase(sg *parse.Graph, sequence string, score ScoreFn) *Base {
	g := new(Base)
	g.Score = score
	g.Labels = make([]rune, len(sg.Nodes))
	copy(g.Labels, sg.Nodes)

	g.SetSeq(sequence)
	g.Edges, g.Loops = normalizeEdges(sg.Edges, len(g.Labels))
	return g
}

func (g *Base) SetSeq(s string) {
	g.SeqLabels = []rune(s)
	g.order = len(g.SeqLabels)*len(g.Labels) + 2
	g.Src = g.order - 2
	g.Dst = g.order - 1
}

func (g *Base) Graph() *Graph {
	return &Graph{
		Interface: g,
		Src:       g.Src,
		Dst:       g.Dst,
		Edges:     g.Edges,
		Labels:    g.Labels,
		SeqLabels: g.SeqLabels,
		Loops:     g.Loops,
		Score:     g.Score,
		order:     g.order,
	}
}

func (g *Base) VisitFromSrc(do func(w int, c int64) bool) bool {
	neighbors := len(g.Labels)
	for i := 0; i < neighbors; i++ {
		c := g.Score(g.SeqLabels[0], g.Labels[i])
		if do(i, c) {
			return true
		}
	}
	return false
}

func (g *Base) VisitFromLastRow(v int, do func(w int, c int64) bool) bool {
	return do(g.Dst, 0)
}

func normalizeEdges(es [][2]int, vertices int) ([][]int, []bool) {
	es = removeDupEdgs(es)
	// outEdgs count how many edges going out of each vertex.
	outEdgs := make([]int, vertices)
	// loops keeps track of which vertex has loops.
	loops := make([]bool, vertices)
	for _, e := range es {
		// Check loop.
		if e[0] == e[1] {
			loops[e[0]] = true
			// Remove loop from the counted edges because
			// it will not be used by the shortest path.
			continue
		}
		outEdgs[e[0]]++
	}
	edges := make([][]int, vertices)
	for i := range edges {
		// l = (horizontal + diagonal) + vertical.
		l := 2*outEdgs[i] + 1
		edges[i] = make([]int, 0, l)
	}
	for _, e := range es {
		// Avoid loops.
		if e[0] == e[1] {
			continue
		}
		// Horizontal edges.
		src, dst := e[0], e[1]
		edges[src] = append(edges[src], dst)
	}
	for i := range edges {
		hLen := len(edges[i])
		// Vertical edge.
		edges[i] = append(edges[i], i+vertices)
		start := len(edges[i])
		// Diagonal edges.
		// Append all the horizontal edges and then sum each
		// one to the vertices to get the diagonal.
		edges[i] = append(edges[i], edges[i][:hLen]...)
		for j := start; j < len(edges[i]); j++ {
			edges[i][j] += vertices
		}
	}
	// The final sequence of edges from each vertex is
	// horizontals:vertical:diagonals.
	return edges, loops
}

func removeDupEdgs(es [][2]int) [][2]int {
	uniques := make(map[[2]int]struct{})
	for _, e := range es {
		uniques[e] = struct{}{}
	}
	edges := make([][2]int, len(uniques))
	i := 0
	for e := range uniques {
		edges[i] = e
		i++
	}
	return edges
}
