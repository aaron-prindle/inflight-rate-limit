package inflight

import (
	"fmt"
)

type sharedQuotaManager struct {
	producers map[PriorityBand]*quotaProducer
}

func newSharedQuotaManager(quotaCh chan<- interface{}, queues []*Queue, qd *queueDrainer) *sharedQuotaManager {
	mgr := &sharedQuotaManager{
		producers: make(map[PriorityBand]*quotaProducer),
	}
	for i := 0; i <= int(SystemLowestPriorityBand); i++ {
		mgr.producers[PriorityBand(i)] = &quotaProducer{
			queues:   queues,
			priority: PriorityBand(i),
			qd:       qd,
		}
	}
	// TODO(aaron-prindle) FIX - this needs to be dynamic...
	for _, prioritylvl := range Priorities {
		mgr.producers[prioritylvl].remainingQuota += ACV(prioritylvl, mgr.producers[prioritylvl].queues)
	}

	return mgr
}

func (m *sharedQuotaManager) Run() {
	for i, producer := range m.producers {
		producer := producer
		fmt.Printf("producer[%d] starting...\n", i)
		go producer.Run()
	}
}
