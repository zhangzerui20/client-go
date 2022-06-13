package treworkqueue

import (
	"k8s.io/client-go/util/workqueue"
	"k8s.io/utils/clock"
	"time"
)

type t interface{}

type treRateLimitQueue struct{
	workqueue.Interface

	rateLimiter workqueue.RateLimiter

	waitingForAddCh chan *waitFor

	clock clock.WithTicker
}

// waitFor holds the data to add and the time it should be added
type waitFor struct {
	data    t
	readyAt time.Time
	// index in the priority queue (heap)
	index int
}

type waitForPriorityQueue []*waitFor

func (pq waitForPriorityQueue) Len() int { return len(pq) }

func (pq waitForPriorityQueue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return pq[i].readyAt.Before(pq[i].readyAt)
}

func (pq waitForPriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *waitForPriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*waitFor)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *waitForPriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

func NewDelayingQueueWithCustomClock(clock clock.WithTicker, name string) workqueue.DelayingInterface {
	return newDelayingQueue(clock, workqueue.NewNamed(name), name)
}

func newDelayingQueue(clock clock.WithTicker, q workqueue.Interface, name string) workqueue.DelayingInterface {
	return treRateLimitQueue{
		Interface: q,
		rateLimiter: workqueue.DefaultControllerRateLimiter(),
		waitingForAddCh: make(chan *waitFor, 1000),
		clock: clock,
	}
}

func (t treRateLimitQueue) AddAfter(item interface{}, duration time.Duration) {



}

func (t treRateLimitQueue) AddRateLimited(item interface{}) {
	panic("implement me")
}

func (t treRateLimitQueue) Forget(item interface{}) {
	panic("implement me")
}

func (t treRateLimitQueue) NumRequeues(item interface{}) int {
	panic("implement me")
}




