package inflight

import (
	"fmt"
	"math"
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
	queues                   []*Queue
	remainingQuota           map[int]int
	remainingQuotaByPriority map[PriorityBand]int
	// remainingQuota map[string]int
}

// AssuredConcurrencyShares is a positive number for a non-exempt priority level, representing the weight by which
// the priority level shares the concurrency from the global limit. The concurrency limit of an apiserver is divided
// among the non-exempt priority levels in proportion to their assured concurrency shares. Basically this produces
// the assured concurrency value (ACV) for each priority level:
//
//             ACV(l) = ceil( SCL * ACS(l) / ( sum[priority levels k] ACS(k) ) )
//

func (c *quotaProducer) ACS(pl PriorityBand) int {
	// priority level -> queues
	// get all queues for the priority level
	// sum the Priority values for those queues
	assuredconcurrencyshares := 0
	for _, queue := range c.queues {
		if queue.Priority == pl {
			assuredconcurrencyshares += queue.SharedQuota
		}
	}
	return assuredconcurrencyshares
}

func (c *quotaProducer) ACV(pl PriorityBand) int {
	// ACV(l) = ceil( SCL * ACS(l) / ( sum[priority levels k] ACS(k) ) )

	// SCL := 400 // SCL is the apiserver's concurrency limit and ACS(l) is the
	denom := 0
	for _, prioritylvl := range Priorities {
		denom += c.ACS(prioritylvl)
	}
	return int(math.Ceil(float64(SCL * c.ACS(pl) / denom)))

}

func (c *quotaProducer) Run(quotaProcessFunc func(queue *Queue, quotaReleaseFunc func())) {
	time.Sleep(1 * time.Second)
	for {
		// fmt.Println("outer quota_producer.go ...")
		for _, prioritylvl := range PriorityBand {
			// for i, queue := range c.queues {
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
