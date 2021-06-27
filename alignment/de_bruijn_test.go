package alignment

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hbollon/go-edlib"
	"github.com/rschio/align/debruijn"
	"github.com/rschio/align/parse"
)

func putErrs1(s string, _ int, percent float64) string {
	return putErrs(s, percent)
}

func putErrs2(s string, k int, percent float64) string {
	var alphabet = []rune("ACGT")
	a := len(alphabet)
	rs := []rune(s)
	sq := make([]rune, len(rs))
	for i := range sq {
		sq[i] = alphabet[rand.Intn(a)]
	}
	copy(sq, rs[:k])
	for i := 0; i < int(percent*float64(k)); i++ {
		pos := rand.Intn(k)
		sq[pos] = alphabet[rand.Intn(a)]
	}
	l := len(sq) - k + 1
	for i := 1; i < l; i++ {
		s0 := sq[i : i+k]
		s1 := rs[i : i+k]
		p := diffPercent(s0, s1)
		if p >= percent {
			sq[i+k-1] = rs[i+k-1]
		}
	}
	return string(sq)
}

func putErrs(s string, percent float64) string {
	var alphabet = []rune("ACGT")
	runes := []rune(s)
	a := len(alphabet)
	n := int(percent * float64(len(runes)))
	del := n / 3
	ins := n / 3
	sub := n / 3
	for i := 0; i < del; i++ {
		l := len(runes)
		p := rand.Intn(l)
		copy(runes[p:], runes[p+1:])
		runes = runes[:len(runes)-1]
	}
	for i := 0; i < ins; i++ {
		l := len(runes)
		p := rand.Intn(l)
		r := rand.Intn(a)
		runes = append(runes[:p], append([]rune{alphabet[r]}, runes[p:]...)...)
	}
	for i := 0; i < sub; i++ {
		l := len(runes)
		p := rand.Intn(l)
		r := rand.Intn(a)
		if alphabet[r] == runes[p] {
			r = (r + 1) % a
		}
		runes[p] = alphabet[r]
	}
	return string(runes)
}

func findAPath(g *parse.Graph, k, length int) ([]rune, []int) {
again:
	path := make([]int, 0, 10)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	v := r.Intn(len(g.Nodes))
	v = v - (v % k)
	i := 0
	for ; i < length; i++ {
		ns := []int{}
		for _, e := range g.Edges {
			if e[0] == v {
				ns = append(ns, e[1])
			}
		}
		if len(ns) == 0 {
			break
		}
		p := r.Intn(len(ns))
		next := ns[p]
		path = append(path, v)
		v = next
	}
	if len(path) < length/2 {
		goto again
	}
	runes := make([]rune, len(path))
	for i, v := range path {
		runes[i] = g.Nodes[v]
	}
	return runes, path
}

func TestNDeBruijn(t *testing.T) {
	const k = 7

	pattern := filepath.Join("testdata", "dbg", "nseqs", "*.txt")
	fnames, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatal(err)
	}

	sep := strings.Repeat("-", k)
	allSeqs := ""
	for _, fname := range fnames {
		seq, err := readSequence(fname)
		if err != nil {
			t.Fatal(err)
		}
		allSeqs = allSeqs + seq + sep
	}
	bg := debruijn.NewDeBruijn([]rune(allSeqs), k+1)
	newEdgs := make([][2]int, 0, len(bg.Edges))
	for _, e := range bg.Edges {
		p0 := strings.Index(string(bg.Vertices[e[0]]), "-")
		p1 := strings.Index(string(bg.Vertices[e[1]]), "-")
		if p0 >= 0 || p1 >= 0 {
			continue
		}
		newEdgs = append(newEdgs, e)
	}
	bg.Edges = newEdgs

	ppg := bg.Parse()
	pg := &parse.Graph{Nodes: ppg.Vertices, Edges: ppg.Edges}

	for i := 0; i < 3; i++ {
		p, _ := findAPath(pg, k, 5000)
		correct := string(p)

		incorrect := putErrs(correct, 0.18)
		b := NewBase(pg, incorrect, weight)
		g := NewDBG(b, k).Graph()
		path, _ := g.ShortestPath()
		s1, t1 := g.Align(path)
		s1, t1 = cmpStr(s1, t1)
		sim, err := edlib.StringsSimilarity(correct, s1, edlib.Levenshtein)
		if err != nil {
			log.Fatal("failed to compare strings")
		}
		t.Logf("%d\tSize: %d\tSimilarity: %v", i, len(p), sim)
	}
}

func TestDeBruijn(t *testing.T) {
	const k = 7
	seqfile := filepath.Join("testdata", "dbg", "seq_5000.txt")
	incorrectfile := filepath.Join("testdata", "dbg", "incorrect_5000.txt")
	seq, err := readSequence(seqfile)
	if err != nil {
		t.Fatal(err)
	}
	incorrect, err := readSequence(incorrectfile)
	if err != nil {
		t.Fatal(err)
	}
	bg := debruijn.NewDeBruijn([]rune(seq), k+1).Parse()
	pg := &parse.Graph{Nodes: bg.Vertices, Edges: bg.Edges}
	b := NewBase(pg, incorrect, weight)
	g := NewDBG(b, k).Graph()
	path, dist := g.ShortestPath()
	s1, t1 := g.Align(path)
	s1, t1 = cmpStr(s1, t1)
	sim, err := edlib.StringsSimilarity(seq, s1, edlib.Levenshtein)
	if err != nil {
		log.Fatal("failed to compare strings")
	}
	t.Log(sim)
	t.Log(dist)
	t.Log(len(s1))
	t.Log(len(t1))
	t.Log(len(seq))
}

type dbgFilterArgs struct {
	GlobalErr bool `json:"global_err"`
	K         int  `json:"k"`
	Vertices  int  `json:"vertices"`
	Filter    bool `json:"filter"`
}

type Metrics struct {
	Duration          time.Duration  `json:"duration"`
	FilterDuration    time.Duration  `json:"filter_duration"`
	AlignmentDuration time.Duration  `json:"alignment_duration"`
	Similarity        float32        `json:"similarity"`
	Distance          int64          `json:"distance"`
	Args              *dbgFilterArgs `json:"args"`
}

func BenchmarkDeBruijnFilterOrNot(b *testing.B) {
	benchs := []dbgFilterArgs{
		{GlobalErr: true, K: 21, Vertices: 10000, Filter: true},
		{GlobalErr: true, K: 21, Vertices: 11000, Filter: true},
		{GlobalErr: true, K: 21, Vertices: 12000, Filter: true},
		{GlobalErr: true, K: 21, Vertices: 13000, Filter: true},
		{GlobalErr: true, K: 21, Vertices: 14000, Filter: true},
		{GlobalErr: true, K: 21, Vertices: 15000, Filter: true},
		{GlobalErr: true, K: 21, Vertices: 10000, Filter: false},
		{GlobalErr: true, K: 21, Vertices: 11000, Filter: false},
		{GlobalErr: true, K: 21, Vertices: 12000, Filter: false},
		{GlobalErr: true, K: 21, Vertices: 13000, Filter: false},
		{GlobalErr: true, K: 21, Vertices: 14000, Filter: false},
		{GlobalErr: true, K: 21, Vertices: 15000, Filter: false},
		{GlobalErr: true, K: 25, Vertices: 10000, Filter: true},
		{GlobalErr: true, K: 25, Vertices: 11000, Filter: true},
		{GlobalErr: true, K: 25, Vertices: 12000, Filter: true},
		{GlobalErr: true, K: 25, Vertices: 13000, Filter: true},
		{GlobalErr: true, K: 25, Vertices: 14000, Filter: true},
		{GlobalErr: true, K: 25, Vertices: 15000, Filter: true},
		{GlobalErr: true, K: 25, Vertices: 10000, Filter: false},
		{GlobalErr: true, K: 25, Vertices: 11000, Filter: false},
		{GlobalErr: true, K: 25, Vertices: 12000, Filter: false},
		{GlobalErr: true, K: 25, Vertices: 13000, Filter: false},
		{GlobalErr: true, K: 25, Vertices: 14000, Filter: false},
		{GlobalErr: true, K: 25, Vertices: 15000, Filter: false},
		{GlobalErr: true, K: 29, Vertices: 10000, Filter: true},
		{GlobalErr: true, K: 29, Vertices: 11000, Filter: true},
		{GlobalErr: true, K: 29, Vertices: 12000, Filter: true},
		{GlobalErr: true, K: 29, Vertices: 13000, Filter: true},
		{GlobalErr: true, K: 29, Vertices: 14000, Filter: true},
		{GlobalErr: true, K: 29, Vertices: 15000, Filter: true},
		{GlobalErr: true, K: 29, Vertices: 10000, Filter: false},
		{GlobalErr: true, K: 29, Vertices: 11000, Filter: false},
		{GlobalErr: true, K: 29, Vertices: 12000, Filter: false},
		{GlobalErr: true, K: 29, Vertices: 13000, Filter: false},
		{GlobalErr: true, K: 29, Vertices: 14000, Filter: false},
		{GlobalErr: true, K: 29, Vertices: 15000, Filter: false},
		{GlobalErr: false, K: 21, Vertices: 10000, Filter: true},
		{GlobalErr: false, K: 21, Vertices: 11000, Filter: true},
		{GlobalErr: false, K: 21, Vertices: 12000, Filter: true},
		{GlobalErr: false, K: 21, Vertices: 13000, Filter: true},
		{GlobalErr: false, K: 21, Vertices: 14000, Filter: true},
		{GlobalErr: false, K: 21, Vertices: 15000, Filter: true},
		{GlobalErr: false, K: 21, Vertices: 10000, Filter: false},
		{GlobalErr: false, K: 21, Vertices: 11000, Filter: false},
		{GlobalErr: false, K: 21, Vertices: 12000, Filter: false},
		{GlobalErr: false, K: 21, Vertices: 13000, Filter: false},
		{GlobalErr: false, K: 21, Vertices: 14000, Filter: false},
		{GlobalErr: false, K: 21, Vertices: 15000, Filter: false},
		{GlobalErr: false, K: 25, Vertices: 10000, Filter: true},
		{GlobalErr: false, K: 25, Vertices: 11000, Filter: true},
		{GlobalErr: false, K: 25, Vertices: 12000, Filter: true},
		{GlobalErr: false, K: 25, Vertices: 13000, Filter: true},
		{GlobalErr: false, K: 25, Vertices: 14000, Filter: true},
		{GlobalErr: false, K: 25, Vertices: 15000, Filter: true},
		{GlobalErr: false, K: 25, Vertices: 10000, Filter: false},
		{GlobalErr: false, K: 25, Vertices: 11000, Filter: false},
		{GlobalErr: false, K: 25, Vertices: 12000, Filter: false},
		{GlobalErr: false, K: 25, Vertices: 13000, Filter: false},
		{GlobalErr: false, K: 25, Vertices: 14000, Filter: false},
		{GlobalErr: false, K: 25, Vertices: 15000, Filter: false},
		{GlobalErr: false, K: 29, Vertices: 10000, Filter: true},
		{GlobalErr: false, K: 29, Vertices: 11000, Filter: true},
		{GlobalErr: false, K: 29, Vertices: 12000, Filter: true},
		{GlobalErr: false, K: 29, Vertices: 13000, Filter: true},
		{GlobalErr: false, K: 29, Vertices: 14000, Filter: true},
		{GlobalErr: false, K: 29, Vertices: 15000, Filter: true},
		{GlobalErr: false, K: 29, Vertices: 10000, Filter: false},
		{GlobalErr: false, K: 29, Vertices: 11000, Filter: false},
		{GlobalErr: false, K: 29, Vertices: 12000, Filter: false},
		{GlobalErr: false, K: 29, Vertices: 13000, Filter: false},
		{GlobalErr: false, K: 29, Vertices: 14000, Filter: false},
		{GlobalErr: false, K: 29, Vertices: 15000, Filter: false},
	}
	datapath := filepath.Join("testdata", "benchdata", "filter_or_not")
	seq, err := readSequence(filepath.Join(datapath, "sequence.txt"))
	if err != nil {
		b.Fatalf("failed to read sequence: %v", err)
	}
	metrics := make([]Metrics, len(benchs))
	for i, bb := range benchs {
		var errfn func(string, int, float64) string
		if bb.GlobalErr {
			errfn = putErrs1
		} else {
			errfn = putErrs2
		}
		seqWithErr := errfn(seq, bb.K, 0.18)
		fname := filepath.Join(datapath, strconv.Itoa(bb.Vertices)+".txt")
		graphSeq, err := readSequence(fname)
		if err != nil {
			b.Fatalf("failed to read graph: %v", err)
		}
		start := time.Now()
		graph := debruijn.NewDeBruijn([]rune(graphSeq), bb.K+1)
		if bb.Filter {
			fstart := time.Now()
			graph.Filter([]rune(seqWithErr), 0.18)
			metrics[i].FilterDuration = time.Since(fstart)
		}
		pgraph := graph.Parse()
		pg := &parse.Graph{Nodes: pgraph.Vertices, Edges: pgraph.Edges}
		b := NewBase(pg, seqWithErr, weight)
		g := NewDBG(b, bb.K).Graph()

		astart := time.Now()
		path, dist := g.ShortestPath()
		s1, t1 := g.Align(path)
		s1, t1 = cmpStr(s1, t1)
		metrics[i].AlignmentDuration = time.Since(astart)
		sim, err := edlib.StringsSimilarity(seq, s1, edlib.Levenshtein)
		if err != nil {
			log.Fatal("failed to compare strings")
		}
		metrics[i].Duration = time.Since(start)
		metrics[i].Similarity = sim
		metrics[i].Distance = dist
		metrics[i].Args = &benchs[i]
	}
	enc := json.NewEncoder(os.Stdout)
	for _, m := range metrics {
		if err := enc.Encode(m); err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkDeBruijnReal(b *testing.B) {
	//ks := []int{21, 25, 29}
	ks := []int{7}
	for j := 0; j < 2; j++ {
		for i := 1; i <= 3; i++ {
			n := strconv.Itoa(i)
			for kn := 0; kn < len(ks); kn++ {
				k := ks[kn]
				datapath := filepath.Join("testdata", "benchdata", "real")
				seq, err := readSequence(filepath.Join(datapath, "sequence"+n+".txt"))
				if err != nil {
					b.Fatalf("failed to read sequence: %v", err)
				}
				fname := filepath.Join(datapath, "graph"+n+".txt")
				graphSeq, err := readSequence(fname)
				graph := debruijn.NewDeBruijn([]rune(graphSeq), k+1)
				if err != nil {
					b.Fatalf("failed to read graph: %v", err)
				}
				graph.FilterGaps('-')
				if j == 0 {
					graph.Filter([]rune(seq), 0.18)
				}
				pgraph := graph.Parse()
				pg := &parse.Graph{Nodes: pgraph.Vertices, Edges: pgraph.Edges}
				bb := NewBase(pg, seq, weight)
				g := NewDBG(bb, k).Graph()
				start := time.Now()
				path, dist := g.ShortestPath()
				s1, t1 := g.Align(path)
				ttime := time.Since(start)
				s1, t1 = cmpStr(s1, t1)
				sim, err := edlib.StringsSimilarity(seq, s1, edlib.Levenshtein)
				if err != nil {
					log.Fatal("failed to compare strings")
				}
				fmt.Printf("Sim: %.2f", sim)
				fmt.Printf("Dist: %d\n", dist)
				fmt.Printf("I: %d\tK: %d\n", i, k)
				fmt.Printf("Dur: %v\n", ttime)
				fmt.Printf("%s\n", s1)
			}
		}
	}
}

func BenchmarkDeBruijnFilter(b *testing.B) {
	const k = 7
	seq, err := readSequence(filepath.Join("testdata", "benchdata",
		"sequence_data", "seq_1500.txt"))
	seqB1, err := readSequence(filepath.Join("testdata", "benchdata",
		"sequence_data", "seq_2000.txt"))
	seqB2, err := readSequence(filepath.Join("testdata", "benchdata",
		"sequence_data", "seq_2500.txt"))
	seqA1, err := readSequence(filepath.Join("testdata", "benchdata",
		"sequence_data", "seq_3000.txt"))
	seqA2, err := readSequence(filepath.Join("testdata", "benchdata",
		"sequence_data", "seq_3500.txt"))
	seqA3, err := readSequence(filepath.Join("testdata", "benchdata",
		"sequence_data", "seq_4000.txt"))
	if err != nil {
		b.Fatal(err)
	}
	incseq := putErrs(seq, 0.18)
	fmt.Println(maxDiff(incseq, seq, k))
	graphseq := seqB1 + seqB2 + seq + seqA1 + seqA2 + seqA3

	bg00 := debruijn.NewDeBruijn([]rune(graphseq), k+1)
	bg0 := bg00.Parse()
	bg01 := debruijn.NewDeBruijn([]rune(graphseq), k+1)

	filterStart := time.Now()
	bg01.Filter([]rune(incseq), 0.18)
	fmt.Printf("Filter time: %v\n", time.Since(filterStart))
	bg1 := bg01.Parse()
	fmt.Println(len(bg0.Vertices))
	fmt.Println(len(bg1.Vertices))
	bgs := [...]*debruijn.ParseGraph{bg0, bg1}
	for _, bg := range bgs {
		start := time.Now()
		pg := &parse.Graph{Nodes: bg.Vertices, Edges: bg.Edges}
		b := NewBase(pg, incseq, weight)
		g := NewDBG(b, k).Graph()
		path, dist := g.ShortestPath()
		s1, t1 := g.Align(path)
		s1, t1 = cmpStr(s1, t1)
		sim, err := edlib.StringsSimilarity(seq, s1, edlib.Levenshtein)
		if err != nil {
			log.Fatal("failed to compare strings")
		}
		elapsed := time.Since(start)
		fmt.Printf("Similarity: %v\n", sim)
		fmt.Printf("Dist: %v\n", dist)
		fmt.Printf("Elapsed: %v\n", elapsed)
	}

}

func diffPercent(s0, s1 []rune) float64 {
	if len(s0) != len(s1) {
		panic("different lengths")
	}
	l := len(s0)
	d, _ := edlib.HammingDistance(string(s0), string(s1))
	return float64(d) / float64(l)
}

func maxDiff(s0, s1 string, k int) float64 {
	if len(s0) != len(s1) {
		panic("different lengths")
	}
	l := len(s0)
	maxglob := 0
	for i := 0; i < l-k+1; i++ {
		maxloc := k
		for j := 0; j < l-k+1; j++ {
			d, _ := edlib.HammingDistance(s0[i:i+k], s1[j:j+k])
			if d < maxloc {
				maxloc = d
			}
		}
		if maxloc > maxglob {
			maxglob = maxloc
		}
	}
	return float64(maxglob) / float64(k)
}

func cmpStr(ss1, ss2 string) (string, string) {
	s1 := []rune(ss1)
	s2 := []rune(ss2)
	if len(s1) > len(s2) {
		s1, s2 = s2, s1
	}
	l := len(s1)
	t1 := make([]rune, 0, l)
	t2 := make([]rune, 0, l)
	for i := range s1 {
		if s1[i] == space {
			continue
		} else if s2[i] == space {
			t1 = append(t1, s1[i])
			t2 = append(t2, s1[i])
			continue
		}
		t1 = append(t1, s1[i])
		t2 = append(t2, s2[i])
	}
	return string(t1), string(t2)
}
