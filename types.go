package inflight

import "time"

type PriorityBand int

const SCL = 400 // SCL is the apiserver's concurrency limit

const (
	SystemTopPriorityBand = PriorityBand(iota)
	SystemHighPriorityBand
	SystemMediumPriorityBand
	SystemNormalPorityBand
	SystemLowPriorityBand

	// This is an implicit priority that cannot be set via API
	SystemLowestPriorityBand
)

var Priorities = []PriorityBand{
	SystemTopPriorityBand,
	SystemHighPriorityBand,
	SystemMediumPriorityBand,
	SystemNormalPorityBand,
	SystemLowPriorityBand,
	SystemLowestPriorityBand,
}

const G = 100000 //   100000 nanoseconds = .1 milliseconds || const G = 60000000000 nanoseconds = 1 minute

// Packet is a temporary container for "requests" with additional tracking fields
// required for the functionality FQScheduler as well as testing
type Packet struct {
	item      interface{}
	size      int
	queueidx  int
	seq       int
	startTime time.Time
}

// Queue is an array of packets with additional metadata required for
// the FQScheduler
type Queue struct {
	Packets           []*Packet
	virstart          float64
	RequestsExecuting int
	Priority          PriorityBand
	SharedQuota       int
	Index             int
}

// Enqueue enqueues a packet into the queue
func (q *Queue) Enqueue(packet *Packet) {
	q.Packets = append(q.Packets, packet)
}

// Dequeue dequeues a packet from the queue
func (q *Queue) Dequeue() (*Packet, bool) {
	if len(q.Packets) == 0 {
		return nil, false
	}
	packet := q.Packets[0]
	q.Packets = q.Packets[1:]
	return packet, true
}

// VirtualFinish returns the expected virtual finish time of the Jth packet in the queue
func (q *Queue) VirtualFinish(J int) float64 {
	// The virtual finish time of request number J in the queue
	// (counting from J=1 for the head) is J * G + (virtual start time).

	J++ // counting from J=1 for the head (eg: queue.Packets[0] -> J=1)
	jg := float64(J) * float64(G)
	return jg + q.virstart
}
