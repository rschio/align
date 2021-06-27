package graph

// ShortestPath computes a shortest path from v to w.
// Only edges with non-negative costs are included.
// The number dist is the length of the path, or -1 if w cannot be reached.
//
// The time complexity is O((|E| + |V|)⋅log|V|), where |E| is the number of edges
// and |V| the number of vertices in the graph.
func ShortestPath(g Iterator, v, w int) (path []int, dist int64) {
	q := &dijkstraQueue{}
	return ShortestPathWithQueue(g, q, v, w)
}

// ShortestPaths computes the shortest paths from v to all other vertices.
// Only edges with non-negative costs are included.
// The number parent[w] is the predecessor of w on a shortest path from v to w,
// or -1 if none exists.
// The number dist[w] equals the length of a shortest path from v to w,
// or is -1 if w cannot be reached.
//
// The time complexity is O((|E| + |V|)⋅log|V|), where |E| is the number of edges
// and |V| the number of vertices in the graph.
func ShortestPaths(g Iterator, v int) (parent []int, dist []int64) {
	n := g.Order()
	dist = make([]int64, n)
	parent = make([]int, n)
	for i := range dist {
		dist[i], parent[i] = -1, -1
	}
	q := &dijkstraQueue{}
	q.SetDist(dist)
	q.Push(v, 0)
	p := &pathFinder{dist: dist, parent: parent, q: q}
	do := p.Do
	for q.Len() > 0 {
		v = q.Pop()
		p.v = v
		g.Visit(v, do)
	}
	return
}

type DistQueue interface {
	// SetDist sets the dist slice to the queue
	// as the priority slice. The queue should
	// use the dist slice as a shared slice, it
	// should not be copied.
	SetDist(dist []int64)
	// Push push v to the queue with
	// cost priority.
	Push(v int, cost int64)
	// Fix changes the cost of v to the
	// new cost priority.
	Fix(v int, cost int64)
	// Pop removes the first element of
	// queue and return it.
	Pop() int
	// Len is the queue's length.
	Len() int
}

func ShortestPathWithQueue(g Iterator, q DistQueue, v, w int) (path []int, dist int64) {
	parent, distances := shortestPathWithQueue(g, q, v, w)
	path, dist = []int{}, distances[w]
	if dist == -1 {
		return
	}
	for v := w; v != -1; v = parent[v] {
		path = append(path, v)
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return
}

func shortestPathWithQueue(g Iterator, q DistQueue, v, w int) (parent []int, dist []int64) {
	n := g.Order()
	dist = make([]int64, n)
	parent = make([]int, n)
	for i := range dist {
		dist[i], parent[i] = -1, -1
	}
	q.SetDist(dist)
	q.Push(v, 0)
	p := &pathFinder{dist: dist, parent: parent, q: q}
	do := p.Do
	for q.Len() > 0 {
		v = q.Pop()
		if v == w {
			return
		}
		p.v = v
		g.Visit(v, do)
	}
	return
}

type pathFinder struct {
	dist   []int64
	parent []int
	q      DistQueue
	v      int
}

func (p *pathFinder) Do(w int, d int64) (skip bool) {
	if d < 0 {
		return
	}
	alt := p.dist[p.v] + d
	switch {
	case p.dist[w] == -1:
		p.parent[w] = p.v
		p.q.Push(w, alt)
	case alt < p.dist[w]:
		p.parent[w] = p.v
		p.q.Fix(w, alt)
	}
	return
}
