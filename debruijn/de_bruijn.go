package debruijn

import (
	"bytes"
	"fmt"
	"strings"
)

type DeBruijn struct {
	Vertices [][]rune
	Edges    [][2]int
	K        int
}

func (g *DeBruijn) String() string {
	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, "digraph \"DeBruijn Graph\" {\n")
	for i, v := range g.Vertices {
		vs := string(v)
		fmt.Fprintf(buf, "\t%d [label=\"%s\"] ;\n", i, vs)
	}
	for _, e := range g.Edges {
		fmt.Fprintf(buf, "\t%d -> %d ;\n", e[0], e[1])
	}
	fmt.Fprintf(buf, "}\n")
	return buf.String()
}

func NewDeBruijn(seq []rune, k int) *DeBruijn {
	if k <= 1 {
		return new(DeBruijn)
	}
	g := new(DeBruijn)
	g.K = k - 1
	n := k - 1
	m := make(map[string]int)

	g.Vertices = append(g.Vertices, seq[:n])
	m[string(seq[:n])] = 0
	l := 1
	prev := 0
	for i := 1; i < len(seq)-n+1; i++ {
		label := seq[i : i+n]
		str := string(label)
		p, ok := m[str]
		if !ok {
			g.Vertices = append(g.Vertices, label)
			m[str] = l
			p = l
			l++
		}
		g.Edges = append(g.Edges, [2]int{prev, p})
		prev = p
	}
	return g
}

type ParseGraph struct {
	Vertices []rune
	Edges    [][2]int
}

func (g *ParseGraph) String() string {
	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, "digraph \"DeBruijn Graph\" {\n")
	for i, v := range g.Vertices {
		vs := string(v)
		fmt.Fprintf(buf, "\t%d [label=\"%s\"] ;\n", i, vs)
	}
	for _, e := range g.Edges {
		fmt.Fprintf(buf, "\t%d -> %d ;\n", e[0], e[1])
	}
	fmt.Fprintf(buf, "}\n")
	return buf.String()
}

func (g *DeBruijn) Parse() *ParseGraph {
	p := new(ParseGraph)
	for i := range g.Vertices {
		for j := 0; j < g.K; j++ {
			p.Vertices = append(p.Vertices, g.Vertices[i][j])
			if j > 0 {
				// Link the inner vertices.
				offset := i * g.K
				p.Edges = append(p.Edges, [2]int{offset + j - 1, offset + j})
			}
		}
	}
	for _, e := range g.Edges {
		u := g.transform(e[0])
		v := g.transform(e[1])
		p.Edges = append(p.Edges, [2]int{u, v})
	}
	return p
}

func (g *DeBruijn) transform(i int) int {
	k := g.K
	return k*i + k - 1
}

type DistanceFn func(a, b []rune) int

// Filter removes vertices where the distance between all slices of k size
// of sequence and a k-mer are greater than the threshold.
func (g *DeBruijn) Filter(seq []rune, threshold float64) {
	g.filter(seq, hamming, threshold)
}

// filter removes vertices where the distance between all slices of k size
// of sequence and a k-mer are greater than the threshold.
func (g *DeBruijn) filter(seq []rune, fn DistanceFn, threshold float64) {
	k := g.K
	l := len(seq) - (k - 1)
	t := int(float64(k) * threshold)
	m := make(map[int]int)
	count := 0
	for i, v := range g.Vertices {
		for j := 0; j < l; j++ {
			dist := fn(v, seq[j:j+k])
			if dist <= t {
				m[i] = count
				count++
				break
			}
		}
	}
	g.remap(m)
}

func (g *DeBruijn) remap(m map[int]int) {
	vtx := make([][]rune, len(m))
	edg := make([][2]int, 0, len(m))
	for prev, new := range m {
		vtx[new] = g.Vertices[prev]
	}
	for _, e := range g.Edges {
		v0 := e[0]
		v1 := e[1]
		nv0, ok0 := m[v0]
		nv1, ok1 := m[v1]
		if ok0 && ok1 {
			edg = append(edg, [2]int{nv0, nv1})
		}
	}
	g.Vertices = vtx
	g.Edges = edg
}

func (g *DeBruijn) FilterGaps(gap rune) {
	m := make(map[int]int)
	count := 0
	for i, v := range g.Vertices {
		if strings.ContainsRune(string(v), gap) {
			continue
		}
		m[i] = count
		count++
	}
	g.remap(m)
}

// hamming returns the hamming distance from a to b,
// it panics if len(a) != len(b).
func hamming(a, b []rune) int {
	if len(a) != len(b) {
		panic("hamming distance cannot compare different length slices")
	}
	l := len(a)
	d := 0
	for i := 0; i < l; i++ {
		if a[i] != b[i] {
			d++
		}
	}
	return d
}
