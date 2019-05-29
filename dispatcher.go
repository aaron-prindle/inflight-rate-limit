package inflight

import (
	"fmt"
	"sync"
)

type Dispatcher struct {
	lock              sync.Mutex
	queues            []*Queue
	requestsexecuting int
	fqScheduler       *FQScheduler
	ACV               int
}

func (c *Dispatcher) Run() {
	// Dispatching is done independently for each priority level. Whenever (1)
	// a non-exempt priority level's number of running requests is below the
	// level's assured concurrency value and (2) that priority level has a
	// non-empty queue, it is time to dispatch another request for service.
	// The Fair Queuing for Server Requests algorithm below is used to pick a
	// non-empty queue at that priority level. Then the request at the head of
	// that queue is dispatched.

	for {
		// c.lock.Lock()
		// defer c.lock.Unlock()

		// Whenever (1) a non-exempt priority level's number of running
		// requests is below the level's assured concurrency value
		fmt.Printf("c.requestsexecuting: %d\n", c.requestsexecuting)
		if c.requestsexecuting < c.ACV {
			// c.quotaUsed++ // TODO(aaron-prindle) this should be done in enqueue?
			// and (2) that priority level has a non-empty queue
			distributionCh, packet := c.fqScheduler.Dequeue()
			// distributionCh is non nil if priority level has a non-empty queue
			if distributionCh != nil {
				go func() {
					fmt.Println("distributed.")
					distributionCh <- func() {
						// these are called after request is delegated
						c.fqScheduler.FinishPacket(packet)
						c.requestsexecuting--
					}
				}()
			}
		}
	}
}
