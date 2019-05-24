package inflight

import (
	"fmt"
	"math"
	"sync"
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
	remainingQuota int
	// remainingQuota map[string]int
	priority PriorityBand
}

func ACS(pl PriorityBand, queues []*Queue) int {
	// priority level -> queues
	// get all queues for the priority level
	// sum the Priority values for those queues
	assuredconcurrencyshares := 0
	for _, queue := range queues {
		if queue.Priority == pl {
			assuredconcurrencyshares += queue.SharedQuota
		}
	}
	return assuredconcurrencyshares
}

func ACV(pl PriorityBand, queues []*Queue) int {
	// ACV(l) = ceil( SCL * ACS(l) / ( sum[priority levels k] ACS(k) ) )
	denom := 0
	for _, prioritylvl := range Priorities {
		denom += ACS(prioritylvl, queues)
	}
	return int(math.Ceil(float64(SCL * ACS(pl, queues) / denom)))
}

func (c *quotaProducer) Run(quotaProcessFunc func(quotaReleaseFunc func())) {
	for {
		gotQuota := false
		func() {
			c.lock.Lock()
			defer c.lock.Unlock()
			fmt.Printf("c.priority: %d, remainingQuota: %d\n", c.priority, c.remainingQuota)
			if c.remainingQuota > 0 {
				c.remainingQuota--
				gotQuota = true
			}
		}()
		if gotQuota {
			quotaProcessFunc(
				func() {
					c.lock.Lock()
					defer c.lock.Unlock()
					c.remainingQuota++
				})
		}

	}
}

func (c *quotaProducer) quotaincrement(i int) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.remainingQuota++
}
