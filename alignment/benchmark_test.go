package alignment

import (
	"fmt"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/rschio/align/alignment/internal/bitbucket"
	"github.com/rschio/align/parse"
	"github.com/rschio/graph"
)

func seqSize(b *testing.B, size int) string {
	seqfile := filepath.Join("testdata", "benchdata", "sequence_data", "seq_"+strconv.Itoa(size)+".txt")
	seq, err := readSequence(seqfile)
	if err != nil {
		b.Fatal(err)
	}
	return seq
}

func readGraphSize(b *testing.B, size int) *parse.Graph {
	graphdir := filepath.Join("testdata", "benchdata", "graph_data")
	graphfile := filepath.Join(graphdir, "graph_"+strconv.Itoa(size)+"v_4d.txt")
	pGraph, err := readSeqGraph(graphfile)
	if err != nil {
		b.Fatalf("failed to read graphfile %s: %v", graphfile, err)
	}
	return pGraph
}

func BenchmarkDial50(b *testing.B) {
	tests := []struct {
		m int
		v int
	}{
		{1000, 10000}, {1500, 10000}, {2000, 10000},
		{2500, 10000}, {3000, 10000}, {3500, 10000},
		{4000, 10000}, {4500, 10000}, {5000, 10000},
		{1000, 20000}, {1500, 20000}, {2000, 20000},
		{2500, 20000}, {3000, 20000}, {3500, 20000},
		{4000, 20000}, {4500, 20000}, {5000, 20000},
		{1000, 30000}, {1500, 30000}, {2000, 30000},
		{2500, 30000}, {3000, 30000}, {3500, 30000},
		{4000, 30000}, {4500, 30000}, {5000, 30000},
		{1000, 40000}, {1500, 40000}, {2000, 40000},
		{2500, 40000}, {3000, 40000}, {3500, 40000},
		{4000, 40000}, {4500, 40000}, {5000, 40000},
		{1000, 50000}, {1500, 50000}, {2000, 50000},
		{2500, 50000}, {3000, 50000}, {3500, 50000},
		{4000, 50000}, {4500, 50000}, {5000, 50000},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%dm%dv", tt.m, tt.v)
		b.Run(name, func(b *testing.B) {
			s := seqSize(b, tt.m)
			pg := readGraphSize(b, tt.v)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				g := NewBase(pg, s, weight).Graph()
				q := new(bitbucket.Queue)
				_, _ = graph.ShortestPathWithQueue(g, q, g.Src, g.Dst)
			}
		})
	}
}
