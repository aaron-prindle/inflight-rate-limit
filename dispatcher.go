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

func (d *Dispatcher) Run() {
	// Dispatching is done independently for each priority level. Whenever (1)
	// a non-exempt priority level's number of running requests is below the
	// level's assured concurrency value and (2) that priority level has a
	// non-empty queue, it is time to dispatch another request for service.
	// The Fair Queuing for Server Requests algorithm below is used to pick a
	// non-empty queue at that priority level. Then the request at the head of
	// that queue is dispatched.

	for {
		func() {
			d.lock.Lock()
			defer d.lock.Unlock()
			// Whenever (1) a non-exempt priority level's number of running
			// requests is below the level's assured concurrency value
			if d.requestsexecuting < d.ACV {
				// and (2) that priority level has a non-empty queue
				distributionCh, packet := d.fqScheduler.Dequeue()
				// distributionCh is non nil if priority level has a non-empty queue
				if distributionCh != nil {
					fmt.Printf("d.requestsexecuting: %d, d.ACV: %d\n", d.requestsexecuting, d.ACV)
					go func() {
						fmt.Println("distributed.")
						distributionCh <- func() {
							// these are called after request is delegated
							d.fqScheduler.FinishPacket(packet)
							d.requestsexecuting--
						}
					}()
				}
			}
		}()
	}
}
