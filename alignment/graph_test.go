package alignment

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/rschio/align/alignment/internal/bitbucket"
	"github.com/rschio/align/parse"
	"github.com/rschio/graph"
)

func findAndPut0(files []string, target string) (ok bool) {
	for i, f := range files {
		if f == target {
			files[0], files[i] = files[i], files[0]
			return true
		}
	}
	return false
}

func weight(a, b rune) int64 {
	switch {
	case a == b:
		return 0
	case a == space || b == space:
		return 1
	}
	return 1
}

func readSeqGraph(fname string) (*parse.Graph, error) {
	gf, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer gf.Close()
	return parse.Parse(gf)
}

func readSequence(fname string) (string, error) {
	bs, err := ioutil.ReadFile(fname)
	if err != nil {
		return "", err
	}
	bs = bytes.TrimSpace(bs)
	return string(bs), nil
}

func TestBigGraph(t *testing.T) {
	seqfile := filepath.Join("testdata", "benchdata", "sequence_data", "seq_100.txt")
	seq, err := readSequence(seqfile)
	if err != nil {
		t.Fatalf("failed to read seqfile %s: %v", seqfile, err)
	}
	graphfile := filepath.Join("testdata", "benchdata", "graph_data", "graph_10000v_4d.txt")
	pg, err := readSeqGraph(graphfile)
	if err != nil {
		t.Fatalf("failed to read seqfile %s: %v", seqfile, err)
	}
	g := NewBase(pg, seq, weight).Graph()
	p, _ := graph.ShortestPath(g, g.Src, g.Dst)
	s1, t1 := g.Align(p)
	if s1 != t1 {
		t.Fatalf("invalid alignment: got %s\nwant %s", s1, t1)
	}
}

func TestGaph(t *testing.T) {
	dirpattern := filepath.Join("testdata", "tests", "*")
	dirs, err := filepath.Glob(dirpattern)
	if err != nil {
		t.Fatalf("failed to glob dirs: %v", err)
	}
	for _, dir := range dirs {
		files, err := filepath.Glob(filepath.Join(dir, "*"))
		if err != nil {
			t.Fatalf("failed to glob files: %v", err)
		}
		if ok := findAndPut0(files, filepath.Join(dir, "seq.txt")); !ok {
			t.Fatal("failed to find sequence file")
		}
		seq, err := readSequence(files[0])
		if err != nil {
			t.Fatalf("failed to read sequence: %v", err)
		}
		files = files[1:]
		for _, tt := range files {
			t.Run(tt, func(t *testing.T) {
				pg, err := readSeqGraph(tt)
				if err != nil {
					t.Fatal(err)
				}
				g := NewBase(pg, seq, weight).Graph()
				q := new(bitbucket.Queue)
				_, dist := graph.ShortestPathWithQueue(g, q, g.Src, g.Dst)
				if dist != 0 {
					t.Fatalf("invalid distance want: %d, got: %d", 0, dist)
				}
			})
		}
	}

}
