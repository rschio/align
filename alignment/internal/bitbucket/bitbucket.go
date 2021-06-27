package bitbucket

import "github.com/yourbasic/bit"

type Queue struct {
	cost    []int64
	buckets [2]*bit.Set
	offset  int64
	length  int
}

func (q *Queue) SetDist(cost []int64) {
	if q == nil {
		*q = Queue{}
	}
	q.cost = cost
	for i := range q.buckets {
		q.buckets[i] = bit.New()
	}
}

func (q *Queue) Len() int { return q.length }

func (q *Queue) Pop() int {
	if q.buckets[0].Empty() {
		q.buckets[0], q.buckets[1] = q.buckets[1], q.buckets[0]
		q.offset++
	}
	m := q.buckets[0].Max()
	q.buckets[0].Delete(m)
	q.length--
	return m
}

func (q *Queue) Fix(v int, cost int64) {
	q.PopV(v)
	q.Push(v, cost)
}

func (q *Queue) Push(v int, cost int64) {
	p := cost - q.offset
	q.buckets[p].Add(v)
	q.cost[v] = cost
	q.length++
}

func (q *Queue) PopV(v int) {
	p := q.cost[v] - q.offset
	q.buckets[p].Delete(v)
	q.length--
}
