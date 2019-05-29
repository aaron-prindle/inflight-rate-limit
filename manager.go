package inflight

import (
	"fmt"

	"k8s.io/utils/clock"
)

type sharedDispatcher struct {
	producers map[PriorityBand]*Dispatcher
}

func queuesForPriority(priority PriorityBand, queues []*Queue) []*Queue {
	// TODO(aaron-prindle) change this to actual impl
	return []*Queue{queues[priority]}
}

func newSharedDispatcher(queues []*Queue) *sharedDispatcher {

	mgr := &sharedDispatcher{
		producers: make(map[PriorityBand]*Dispatcher),
	}

	clock := clock.RealClock{}
	for _, priority := range Priorities {
		mgr.producers[priority] = &Dispatcher{
			queues:      queues,
			ACV:         ACV(priority, queues),
			fqScheduler: NewFQScheduler(queuesForPriority(priority, queues), clock),
		}
	}
	// TODO(aaron-prindle) FIX - this eventually needs to be dynamic...
	for _, priority := range Priorities {
		mgr.producers[priority].ACV += ACV(priority, mgr.producers[priority].queues)
	}

	return mgr
}

func (m *sharedDispatcher) Run() {
	for i, producer := range m.producers {
		producer := producer
		fmt.Printf("producer[%d] starting...\n", i)
		go producer.Run()
	}
}
