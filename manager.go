package inflight

import (
	"sync"
)

func newSharedQuotaManager(quotaCh chan<- interface{}, queues []*Queue) *sharedQuotaManager {
	mgr := &sharedQuotaManager{
		producers: make(map[PriorityBand]*quotaProducer),
		quotaCh:   quotaCh,
	}
	for i := 0; i <= int(SystemLowestPriorityBand); i++ {
		mgr.producers[PriorityBand(i)] = &quotaProducer{
			lock:           &sync.Mutex{},
			remainingQuota: make(map[int]int),
			queues:         queues,
		}
	}
	for i, queue := range queues {
		mgr.producers[queue.Priority].remainingQuota[i] += queue.SharedQuota
		// mgr.producers[queue.Priority].queueByName[queue.Name] = queue
	}
	return mgr
}

type sharedQuotaManager struct {
	producers map[PriorityBand]*quotaProducer
	quotaCh   chan<- interface{}
}

func (m *sharedQuotaManager) Run() {
	for _, producer := range m.producers {
		producer := producer
		go producer.Run(func(queue *Queue, quotaReleaseFunc func()) {
			m.quotaCh <- sharedQuotaNotification{
				quotaReleaseFunc: quotaReleaseFunc,
			}
		})
	}
}
