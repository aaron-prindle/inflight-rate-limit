package inflight

import (
	"fmt"
	"sync"

	"k8s.io/utils/clock"
)

func newQueueDrainer(queues []*Queue, quotaCh <-chan interface{}) *queueDrainer {
	fqSchedulerByPriority := make(map[PriorityBand]*FQScheduler)

	ACVByPriority := make(map[PriorityBand]int)
	for _, priority := range Priorities {
		ACVByPriority[priority] += ACV(priority, queues)
	}

	for i := 0; i <= int(SystemLowestPriorityBand); i++ {
		fqSchedulerByPriority[PriorityBand(i)] = NewFQScheduler([]*Queue{queues[i]}, clock.RealClock{})
	}

	drainer := &queueDrainer{
		quotaCh:               quotaCh,
		fqSchedulerByPriority: fqSchedulerByPriority,
		ACVByPriority:         ACVByPriority,
	}
	return drainer
}

type queueDrainer struct {
	lock                  sync.Mutex
	queueLength           int
	quotaCh               <-chan interface{}
	fqSchedulerByPriority map[PriorityBand]*FQScheduler
	ACVByPriority         map[PriorityBand]int
}

func (d *queueDrainer) Enqueue(queue *Queue) <-chan func() {
	d.lock.Lock()
	defer d.lock.Unlock()
	fmt.Printf("d.queueLength: %d\n", d.queueLength)
	fmt.Printf("d.ACVByPriority: %d\n", d.ACVByPriority)
	if d.queueLength > d.ACVByPriority[queue.Priority] {
		fmt.Println("d.queueLength > d.ACVByPriority: true")
		return nil
	}
	// Prioritizing
	d.queueLength++
	distributionCh := make(chan func(), 1)
	pkt := &Packet{
		queueidx: 0,
	}
	d.fqSchedulerByPriority[queue.Priority].Enqueue(pkt, distributionCh)
	return distributionCh
}
