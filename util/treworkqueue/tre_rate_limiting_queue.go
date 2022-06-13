package treworkqueue

import (
	"k8s.io/client-go/util/workqueue"
	"k8s.io/utils/clock"
	"time"
)

type treRateLimitQueue struct{
	workqueue.Interface

	rateLimiter workqueue.RateLimiter

	waitingForAddCh chan struct{}

	clock clock.WithTicker
}

func NewDelayingQueueWithCustomClock(clock clock.WithTicker, name string) workqueue.DelayingInterface {
	return newDelayingQueue(clock, NewNamed(name), name)
}

func newDelayingQueue() workqueue.DelayingInterface {

}

func (t treRateLimitQueue) AddAfter(item interface{}, duration time.Duration) {
	panic("implement me")
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




