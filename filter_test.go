package inflight

import (
	"log"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"fmt"
)

// InitQueuesPriority is a convenience method for initializing an array of n queues
// for the full list of priorities
func InitQueuesPriority() []*Queue {
	queues := make([]*Queue, 0, len(Priorities))

	for i, priority := range Priorities {
		queues = append(queues, &Queue{
			Packets:     []*Packet{},
			Priority:    priority,
			SharedQuota: 10,
			Index:       i,
		})
	}
	return queues
}

// test ideas
// 5 flows
// 1 concurrency for each
// make requests take 1 second
// send like 10 requests to each
// verify that one in each level was consumed

func TestInflight(t *testing.T) {
	var count int64
	queues := InitQueuesPriority()
	fqFilter := NewFQFilter(queues)
	fqFilter.Delegate = func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&count, int64(1))
	}
	fqFilter.Run()

	http.HandleFunc("/", fqFilter.Serve)
	fmt.Printf("Serving at 127.0.0.1:8080\n")
	go func() {
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()
	// server := httptest.NewServer(http.HandlerFunc(handler))
	// defer server.Close()

	time.Sleep(1 * time.Second)
	header := make(http.Header)
	header.Add("PRIORITY", "0")
	// TODO(aaron-prindle) make this not localhost 8080...
	req, _ := http.NewRequest("GET", "http://localhost:8080", nil)
	// req, _ := http.NewRequest("GET", server.URL, nil)
	req.Header = header

	w := &Work{
		Request: req,
		N:       20,
		C:       2,
	}
	w.Run()
	// fmt.Printf("results: %v\n", w.results)
	i := 0
	for result := range w.results {
		fmt.Printf("result %d:\n%v\n", i, result)
		i++
	}
	if count != 20 {
		t.Errorf("Expected to send 20 requests, found %v", count)
	}
}

// type flowDesc struct {
// 	// In
// 	ftotal int // Total units in flow
// 	imin   int // Min Packet size
// 	imax   int // Max Packet size

// 	// Out
// 	idealPercent  float64
// 	actualPercent float64
// }

// func genFlow(fq *FQScheduler, desc *flowDesc, key int) {
// 	for i, t := int(1), int(0); t < desc.ftotal; i++ {
// 		// req, err := http.NewRequest("GET", "localhost:8080", nil)
// 		// if err != nil {
// 		// 	panic("nooo")
// 		// }
// 		// req.Header.Set("PRIORITY", strconv.Itoa(key))
// 		// client := http.Client{}
// 		// resp, err := client.Do(req)
// 		// if err != nil {
// 		// 	panic("nooo")
// 		// }
// 		it := new(Packet)
// 		it.queueidx = key
// 		if desc.imin == desc.imax {
// 			it.size = desc.imax
// 		} else {
// 			it.size = desc.imin + rand.Intn(desc.imax-desc.imin)
// 		}
// 		if t+it.size > desc.ftotal {
// 			it.size = desc.ftotal - t
// 		}
// 		t += it.size
// 		it.seq = i
// 		// new packet
// 		fq.Enqueue(it)
// 	}
// }

// func consumeQueue(t *testing.T, fq *FQScheduler, descs []flowDesc) (float64, error) {
// 	active := make(map[int]bool)
// 	var total int
// 	acnt := make(map[int]int)
// 	cnt := make(map[int]int)
// 	seqs := make(map[int]int)

// 	wsum := uint64(len(descs))

// 	for i, ok := fq.Dequeue(); ok; i, ok = fq.Dequeue() {
// 		// callback to update virtualtime w/ correct service time for request
// 		fq.FinishPacket(i)

// 		it := i
// 		seq := seqs[it.queueidx]
// 		if seq+1 != it.seq {
// 			return 0, fmt.Errorf("Packet for flow %d came out of queue out-of-order: expected %d, got %d", it.queueidx, seq+1, it.seq)
// 		}
// 		seqs[it.queueidx] = it.seq

// 		// set the flow this item is a part of to active
// 		if cnt[it.queueidx] == 0 {
// 			active[it.queueidx] = true
// 		}
// 		cnt[it.queueidx] += it.size

// 		// if # of active flows is equal to the # of total flows, add to total
// 		if len(active) == len(descs) {
// 			acnt[it.queueidx] += it.size
// 			total += it.size
// 		}

// 		// if all items have been processed from the flow, remove it from active
// 		if cnt[it.queueidx] == descs[it.queueidx].ftotal {
// 			delete(active, it.queueidx)
// 		}
// 	}

// 	if total == 0 {
// 		t.Fatalf("expected 'total' to be nonzero")
// 	}

// 	var variance float64
// 	for key := 0; key < len(descs); key++ {
// 		// flows in this test have same expected # of requests
// 		// idealPercent = total-all-active/len(flows) / total-all-active
// 		// "how many bytes/requests you expect for this flow - all-active"
// 		descs[key].idealPercent = float64(100) / float64(wsum)

// 		// actualPercent = requests-for-this-flow-all-active / total-reqs
// 		// "how many bytes/requests you got for this flow - all-active"
// 		descs[key].actualPercent = (float64(acnt[key]) / float64(total)) * 100

// 		x := descs[key].idealPercent - descs[key].actualPercent
// 		x *= x
// 		variance += x
// 	}
// 	variance /= float64(len(descs))

// 	stdDev := math.Sqrt(variance)
// 	return stdDev, nil
// }
