package inflight

import (
	"fmt"
	"math"
	"sync"
	"time"

	"k8s.io/utils/clock"
)

// FQScheduler is a fair queuing implementation designed for the kube-apiserver.
// FQ is designed for
// 1) dispatching requests to be served rather than packets to be transmitted
// 2) serving multiple requests at once
// 3) accounting for unknown and varying service time
type FQScheduler struct {
	lock         sync.Mutex
	queues       []*Queue
	clock        clock.Clock
	vt           float64
	C            float64
	G            float64
	lastRealTime time.Time
	robinidx     int
	// --
	queuedistchs map[int][]interface{}
}

func (q *FQScheduler) chooseQueue(packet *Packet) *Queue {
	if packet.queueidx < 0 || packet.queueidx > len(q.queues) {
		panic("no matching queue for packet")
	}
	return q.queues[packet.queueidx]
}

func NewFQScheduler(queues []*Queue, clock clock.Clock) *FQScheduler {
	fq := &FQScheduler{
		queues:       queues,
		clock:        clock,
		vt:           0,
		queuedistchs: make(map[int][]interface{}),
	}
	return fq
}

func (q *FQScheduler) nowAsUnixNano() float64 {
	return float64(q.clock.Now().UnixNano())
}

func (q *FQScheduler) Enqueue(queue *Queue) <-chan func() {
	distributionCh := make(chan func(), 1)
	pkt := &Packet{
		queueidx: queue.Index,
	}
	// TODO(aaron-prindle) make it so enqueue fails if the queues are full
	q.enqueue(pkt, distributionCh)
	return distributionCh
}

// Enqueue enqueues a packet into the fair queuing scheduler
func (q *FQScheduler) enqueue(packet *Packet, distributionCh chan<- func()) {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.synctime()

	queue := q.chooseQueue(packet)
	queue.Enqueue(packet)
	q.updateTime(packet, queue)
	// --
	q.queuedistchs[packet.queueidx] = append(q.queuedistchs[packet.queueidx], distributionCh)
}

func (q *FQScheduler) getVirtualTime() float64 {
	return q.vt
}

func (q *FQScheduler) synctime() {
	realNow := q.clock.Now()
	timesincelast := realNow.Sub(q.lastRealTime).Nanoseconds()
	q.lastRealTime = realNow
	q.vt += float64(timesincelast) * q.getvirtualtimeratio()
}

func (q *FQScheduler) getvirtualtimeratio() float64 {
	NEQ := 0
	reqs := 0
	for _, queue := range q.queues {
		reqs += queue.RequestsExecuting
		if len(queue.Packets) > 0 || queue.RequestsExecuting > 0 {
			NEQ++
		}
	}
	// no active flows, virtual time does not advance (also avoid div by 0)
	if NEQ == 0 {
		return 0
	}
	return min(float64(reqs), float64(SCL)) / float64(NEQ)
}

func (q *FQScheduler) updateTime(packet *Packet, queue *Queue) {
	// When a request arrives to an empty queue with no requests executing
	// len(queue.Packets) == 1 as enqueue has just happened prior (vs  == 0)
	if len(queue.Packets) == 1 && queue.RequestsExecuting == 0 {
		// the queue’s virtual start time is set to getVirtualTime().
		queue.virstart = q.getVirtualTime()
	}
}

// FinishPacket is a callback that should be used when a previously dequeud packet
// has completed it's service.  This callback updates imporatnt state in the
//  FQScheduler
func (q *FQScheduler) FinishPacket(p *Packet) {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.synctime()
	S := q.clock.Since(p.startTime).Nanoseconds()

	// When a request finishes being served, and the actual service time was S,
	// the queue’s virtual start time is decremented by G - S.
	q.queues[p.queueidx].virstart -= G - float64(S)

	// request has finished, remove from requests executing
	q.queues[p.queueidx].RequestsExecuting--
}

// Dequeue dequeues a packet from the fair queuing scheduler
func (q *FQScheduler) Dequeue() (distributionCh chan<- func(), packet *Packet) {
	q.lock.Lock()
	defer q.lock.Unlock()
	queue := q.selectQueue()

	if queue == nil {
		return nil, nil
		// return nil, false
	}
	packet, ok := queue.Dequeue()

	if ok {
		// When a request is dequeued for service -> q.virstart += G
		queue.virstart += G

		packet.startTime = q.clock.Now()
		// request dequeued, service has started
		queue.RequestsExecuting++
	} else {
		return nil, nil
	}
	// dequeue
	id := packet.queueidx
	fmt.Printf("queue: %v\n", queue)
	fmt.Printf("packet: %v\n", packet)
	fmt.Printf("len(q.queuedistchs): %d\n", len(q.queuedistchs))
	distributionCh = q.queuedistchs[id][0].(chan<- func())
	q.queuedistchs[id] = q.queuedistchs[id][1:]
	return distributionCh, packet
}

func (q *FQScheduler) roundrobinqueue() int {
	q.robinidx = (q.robinidx + 1) % len(q.queues)
	return q.robinidx
}

func (q *FQScheduler) selectQueue() *Queue {
	minvirfinish := math.Inf(1)
	var minqueue *Queue
	var minidx int
	for range q.queues {
		idx := q.roundrobinqueue()
		queue := q.queues[idx]
		if len(queue.Packets) != 0 {
			curvirfinish := queue.VirtualFinish(0)
			if curvirfinish < minvirfinish {
				minvirfinish = curvirfinish
				minqueue = queue
				minidx = q.robinidx
			}
		}
	}
	q.robinidx = minidx
	return minqueue
}
