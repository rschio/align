package alignment

func (g *Graph) Align(path []int) (string, string) {
	s := make([]rune, len(path)-2)
	t := make([]rune, len(path)-2)
	rowLen := len(g.Labels)
	// Ignore the fake nodes.
	path = path[1 : len(path)-1]
	seqn := g.SeqLabels
	s[0] = g.Labels[path[0]]
	t[0] = seqn[0]
	seqn = seqn[1:]
	prev := path[0]
	j := 0
	for i := 1; i < len(path); i++ {
		v := path[i]
		if v == prev+rowLen {
			// Vertical.
			s[i] = space
			t[i] = seqn[j]
			j++
		} else if prev/rowLen == v/rowLen {
			// Horizontal.
			s[i] = g.Labels[v%rowLen]
			t[i] = space
		} else {
			// Diagonal.
			s[i] = g.Labels[v%rowLen]
			t[i] = seqn[j]
			j++
		}
		prev = v
	}
	return string(s), string(t)
}
