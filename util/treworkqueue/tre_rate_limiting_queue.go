package treworkqueue

import (
	"container/heap"
	"fmt"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/utils/clock"
	"sync"
	"time"
)

type tt interface{}

type treRateLimitQueue struct {
	workqueue.Interface

	rateLimiter workqueue.RateLimiter

	waitingForAddCh chan *waitFor

	clock clock.WithTicker

	stopCh chan struct{}

	stopOnce sync.Once
}

// waitFor holds the data to add and the time it should be added
type waitFor struct {
	data    tt
	readyAt time.Time
	// index in the priority queue (heap)
	index int
}

type waitForPriorityQueue []*waitFor

func (pq waitForPriorityQueue) Len() int {
	return len(pq)
}
func (pq waitForPriorityQueue) Less(i, j int) bool {
	return pq[i].readyAt.Before(pq[j].readyAt)
}
func (pq waitForPriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

// Push adds an item to the queue. Push should not be called directly; instead,
// use `heap.Push`.
func (pq *waitForPriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*waitFor)
	item.index = n
	*pq = append(*pq, item)
}

// Pop removes an item from the queue. Pop should not be called directly;
// instead, use `heap.Pop`.
func (pq *waitForPriorityQueue) Pop() interface{} {
	n := len(*pq)
	item := (*pq)[n-1]
	item.index = -1
	*pq = (*pq)[0:(n - 1)]
	return item
}

// Peek returns the item at the beginning of the queue, without removing the
// item or otherwise mutating the queue. It is safe to call directly.
func (pq waitForPriorityQueue) Peek() interface{} {
	return pq[0]
}

func NewDelayingQueueWithCustomClock(clock clock.WithTicker, name string) workqueue.DelayingInterface {
	return newDelayingQueue(clock, workqueue.NewNamed(name), name)
}

func newDelayingQueue(clock clock.WithTicker, q workqueue.Interface, name string) *treRateLimitQueue {

	queue := treRateLimitQueue{
		Interface:       q,
		rateLimiter:     workqueue.DefaultControllerRateLimiter(),
		waitingForAddCh: make(chan *waitFor, 1000),
		clock:           clock,
	}

	go queue.waitLoop()

	return &queue

}

func (t treRateLimitQueue) ShutDown() {
	t.stopOnce.Do(func() {
		t.Interface.ShutDown()
		close(t.stopCh)
	})
}

func (t treRateLimitQueue) AddAfter(item interface{}, duration time.Duration) {

	if duration <= 0 {
		return
	}

	w := &waitFor{
		data:    item,
		readyAt: t.clock.Now().Add(duration),
	}

	select {
	case <-t.stopCh:
		return
	case t.waitingForAddCh <- w:
	}

	return
}

func (t *treRateLimitQueue) waitLoop() {

	waitQueue := &waitForPriorityQueue{}
	heap.Init(waitQueue)
	neverOccur := make(<-chan time.Time)
	var nextReadyAtTimer clock.Timer

	waitingEntryByData := make(map[tt]*waitFor)

	for {
		// scan the heap, find those can en-queue
		now := t.clock.Now()
		for waitQueue.Len() > 0 {
			w := waitQueue.Peek().(*waitFor)
			fmt.Printf("judege if enqueue ready at %s, now is %s \n", w.readyAt.String(), now.String())
			if w.readyAt.After(now) {
				// the heap top is not ready, so the rest can't en-queue, too
				break
			}

			w = heap.Pop(waitQueue).(*waitFor)
			t.Add(w.data)
			delete(waitingEntryByData, w.data)
		}

		// now get the next
		nextOccur := neverOccur

		if waitQueue.Len() > 0 {
			w := waitQueue.Peek().(*waitFor)
			if nextReadyAtTimer != nil {
				nextReadyAtTimer.Stop()
			}
			nextReadyAtTimer = t.clock.NewTimer(w.readyAt.Sub(now))
			nextOccur = nextReadyAtTimer.C()
		}

		select {
		case <-t.stopCh:
			return
		case <-nextOccur:
			fmt.Println("next occur happen " + t.clock.Now().String())
		case w := <-t.waitingForAddCh:
			// push in heap
			w.index = waitQueue.Len()
			fmt.Println("push heap, ready at " + w.readyAt.String())
			insert(waitQueue, waitingEntryByData, w)
		}
	}

}

// insert adds the entry to the priority queue, or updates the readyAt if it already exists in the queue
func insert(q *waitForPriorityQueue, knownEntries map[tt]*waitFor, entry *waitFor) {
	// if the entry already exists, update the time only if it would cause the item to be queued sooner
	existing, exists := knownEntries[entry.data]
	if exists {
		if existing.readyAt.After(entry.readyAt) {
			existing.readyAt = entry.readyAt
			heap.Fix(q, existing.index)
		}

		return
	}

	heap.Push(q, entry)
	knownEntries[entry.data] = entry
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
