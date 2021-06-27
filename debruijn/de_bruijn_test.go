package debruijn

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDeBruijn(t *testing.T) {
	seq := []rune("ACGCGTCG")
	g := NewDeBruijn(seq, 3)
	p := g.Parse()
	if len(p.Vertices) != 2*len(g.Vertices) {
		t.Errorf("parse graph should have k*vertices")
	}
}

func TestFilter(t *testing.T) {
	graphseq := []rune("AAABCDEFGHIJKLMNOPQRSTUVWXYZ")
	k := 5
	tests := []struct {
		name      string
		seq       []rune
		want      int
		threshold float64
	}{
		{name: "1", seq: []rune("AAAAA"), want: 1, threshold: 0.25},
		{name: "2", seq: []rune("ABCDEBBBBB"), want: 2, threshold: 0},
		{name: "3", seq: []rune("AAAAAA"), want: 0, threshold: 0},
		{name: "4", seq: []rune("111111"), want: len(graphseq) - k + 2, threshold: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewDeBruijn(graphseq, k)
			g.Filter(tt.seq, tt.threshold)
			got := len(g.Vertices)
			if got != tt.want {
				t.Errorf("want %d, got %d vertices", tt.want, got)
			}
		})
	}
}

func TestReduction(t *testing.T) {
	k := 21
	graphseq := readseq(t, filepath.Join("testdata", "seq_1000000.txt"))
	seq := readseq(t, filepath.Join("testdata", "seq_1500.txt"))
	g := NewDeBruijn(graphseq, k)
	t.Logf("Vertices: %d\tEdges: %d\n", len(g.Vertices), len(g.Edges))
	g.Filter(seq, 0.5)
	t.Logf("Vertices: %d\tEdges: %d\n", len(g.Vertices), len(g.Edges))
}

func TestDBGConnect(t *testing.T) {
	gseq := readseq(t, filepath.Join("testdata", "graph3.txt"))
	k := 7
	g := NewDeBruijn(gseq, k)
	g.FilterGaps('-')
	dedup(g)
	edgs := g.Edges
	in := make([]int, len(g.Vertices))
	out := make([]int, len(g.Vertices))
	for _, edg := range edgs {
		a, b := edg[0], edg[1]
		out[a]++
		in[b]++
	}
	max := 0
	maxi := 0
	avg := 0
	for i := range in {
		t := in[i] + out[i]
		avg += t
		if t > max {
			max = t
			maxi = i
		}
	}
	avg /= len(in)
	t.Logf("Max: %d\n", max)
	t.Logf("Maxi: %d\n", maxi)
	t.Logf("Avg: %d\n", avg)
}

func dedup(g *DeBruijn) {
	edgs := g.Edges
	m := make(map[[2]int]struct{}, len(edgs)/5)
	for _, e := range edgs {
		m[e] = struct{}{}
	}
	newEdg := make([][2]int, 0, len(m))
	for e := range m {
		newEdg = append(newEdg, e)
	}
	g.Edges = newEdg
}

func readseq(t *testing.T, fname string) []rune {
	data, err := os.ReadFile(fname)
	if err != nil {
		t.Fatalf("failed to read seq %s: %v", fname, err)
	}
	return []rune(string(data))
}
