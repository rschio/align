package alignment

type DBG struct {
	*Base
	K int
}

func NewDBG(g *Base, k int) *DBG {
	return &DBG{
		Base: g,
		K:    k,
	}
}

func (g *DBG) Graph() *Graph {
	nGraph := g.Base.Graph()
	nGraph.Interface = g
	return nGraph

}

func (g *DBG) VisitFromSrc(do func(w int, c int64) bool) bool {
	n := len(g.Labels)
	k := g.K
	for i := 0; i < n; i += k {
		c := g.Score(g.SeqLabels[0], g.Labels[i])
		if do(i, c) {
			return true
		}
	}
	return false
}

func (g *DBG) VisitFromLastRow(v int, do func(w int, c int64) bool) bool {
	vertices := len(g.Labels)
	vi := v % vertices
	offset := v - vi
	vertical := vi + vertices
	for _, w := range g.Edges[vi] {
		if w == vertical {
			break
		}
		w += offset
		c := g.Score('A', space)
		if do(w, c) {
			return true
		}
	}
	if v%g.K == g.K-1 {
		return do(g.Dst, 0)
	}
	return false
}
