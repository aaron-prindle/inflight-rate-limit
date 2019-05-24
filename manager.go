package inflight

import (
	"fmt"
	"sync"
)

// quota producer per prioritylevel
// concurrency is per level, not bucket

func newSharedQuotaManager(quotaCh chan<- interface{}, queues []*Queue) *sharedQuotaManager {
	mgr := &sharedQuotaManager{
		producers: make(map[PriorityBand]*quotaProducer),
		quotaCh:   quotaCh,
	}
	for i := 0; i <= int(SystemLowestPriorityBand); i++ {
		mgr.producers[PriorityBand(i)] = &quotaProducer{
			lock:     &sync.Mutex{},
			queues:   queues,
			priority: PriorityBand(i),
		}
	}

	// TODO(aaron-prindle) FIX - this needs to be dynamic...
	for _, prioritylvl := range Priorities {
		mgr.producers[prioritylvl].remainingQuota += ACV(prioritylvl, mgr.producers[prioritylvl].queues)
	}

	return mgr
}

type sharedQuotaManager struct {
	producers map[PriorityBand]*quotaProducer
	quotaCh   chan<- interface{}
}

func (m *sharedQuotaManager) Run() {
	for i, producer := range m.producers {
		producer := producer
		fmt.Printf("producer[%d] starting...\n", i)
		go producer.Run(func(quotaReleaseFunc func()) {
			m.quotaCh <- sharedQuotaNotification{
				priority:         producer.priority,
				quotaReleaseFunc: quotaReleaseFunc,
			}
		})
	}
}
