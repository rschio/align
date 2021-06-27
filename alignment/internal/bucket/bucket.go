// NOTE: This package makes strong assumptions and is wrong for most cases.
// It should only be used for the shortest path algorithm and using a weight
// function that can only return values 0 or 1.
package bucket

type Queue struct {
	index   []int
	cost    []int64
	buckets [2][]int
	offset  int64
	length  int
}

func (q *Queue) SetDist(cost []int64) {
	if q == nil {
		*q = Queue{}
	}
	n := len(cost)
	q.index = make([]int, n)
	q.cost = cost
	for i := range q.buckets {
		q.buckets[i] = make([]int, 0, 500)
	}
}

func (q *Queue) Len() int { return q.length }
func (q *Queue) Fix(v int, cost int64) {
	q.PopV(v)
	q.Push(v, cost)
}

func (q *Queue) Pop() int {
	if len(q.buckets[0]) == 0 {
		q.buckets[0], q.buckets[1] = q.buckets[1], q.buckets[0][:0]
		q.offset++
	}
	b := q.buckets[0]
	n := len(b)
	v := b[n-1]
	q.buckets[0] = b[0 : n-1]
	q.length--
	return v
}

func (q *Queue) Push(v int, cost int64) {
	p := cost - q.offset
	b := q.buckets[p]
	i := len(b)
	q.buckets[p] = append(b, v)
	q.index[v] = i
	q.cost[v] = cost
	q.length++
}

func (q *Queue) PopV(v int) {
	p := q.cost[v] - q.offset
	i := q.index[v]
	b := q.buckets[p]
	n := len(b)
	b[i] = b[n-1]
	q.index[b[n-1]] = i
	q.buckets[p] = b[0 : n-1]
	q.length--
}
