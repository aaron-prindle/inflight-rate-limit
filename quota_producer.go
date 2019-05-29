package inflight

import (
	"fmt"
	"sync"
)

type quotaProducer struct {
	lock           sync.Mutex
	queues         []*Queue
	remainingQuota int
	priority       PriorityBand
	qd             *queueDrainer
}

func (c *quotaProducer) Run() {
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
			func() {
				c.lock.Lock()
				defer c.lock.Unlock()
				fqScheduler := c.qd.fqSchedulerByPriority[c.priority]
				distributionCh, packet := fqScheduler.Dequeue()
				if distributionCh != nil {
					go func() {
						fmt.Println("distributed.")
						distributionCh <- func() { fqScheduler.FinishPacket(packet) }
					}()
					func() {
						c.qd.lock.Lock()
						defer c.qd.lock.Unlock()
						c.qd.queueLength--
						return
					}()
				}
				c.remainingQuota++
			}()
		}
	}
}
