package inflight

import (
	"fmt"
	"sync"
	"time"
)

type sharedQuotaNotification struct {
	priority         PriorityBand
	quotaReleaseFunc func()
}

type reservedQuotaNotification struct {
	priority         PriorityBand
	queueName        string
	quotaReleaseFunc func()
}

type quotaProducer struct {
	lock *sync.Mutex
	// queueByName      map[string]*Queue
	queues         []*Queue
	remainingQuota map[int]int
	// remainingQuota map[string]int
}

func (c *quotaProducer) Run(quotaProcessFunc func(queue *Queue, quotaReleaseFunc func())) {
	time.Sleep(1 * time.Second)
	for {
		// fmt.Println("outer quota_producer.go ...")
		for i, queue := range c.queues {
			queue := queue
			gotQuota := false
			func() {
				c.lock.Lock()
				defer c.lock.Unlock()
				fmt.Printf("c.remainingQuota[%d]: %d\n", i, c.remainingQuota[i])
				if c.remainingQuota[i] > 0 {
					// fmt.Println("reserving quota...")
					c.remainingQuota[i]--
					gotQuota = true
				}
			}()
			if gotQuota {
				quotaProcessFunc(
					queue,
					func() {
						c.lock.Lock()
						defer c.lock.Unlock()
						// fmt.Println("adding back quota...")
						c.remainingQuota[i]++
					})
			}
		}

	}
}

func (c *quotaProducer) quotaincrement(i int) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.remainingQuota[i]++
}
