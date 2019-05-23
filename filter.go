package inflight

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type FQFilter struct {
	// list-watching API models
	lock   *sync.Mutex
	queues []*Queue
	vt     uint64
	C      uint64
	G      uint64
	seen   bool

	// running components
	// *reservedQuotaManager
	*sharedQuotaManager
	*queueDrainer

	Delegate http.HandlerFunc
}

func findMatchedQueue(req *http.Request, queues []*Queue) *Queue {
	index := req.Header.Get("INDEX")
	idx, err := strconv.Atoi(index)
	if err != nil {
		panic("strconv.Atoi(index)")
	}
	return queues[idx]
}

func (f *FQFilter) Serve(resp http.ResponseWriter, req *http.Request) {

	// 0. Matching request w/ bindings API
	matchedQueue := findMatchedQueue(req, f.queues)

	// 1. Waiting to be notified by either a reserved quota or a shared quota
	distributionCh := f.queueDrainer.Enqueue(matchedQueue)
	if distributionCh == nil {
		// too many requests
		fmt.Println("throttled...")
		resp.WriteHeader(409)
	}

	// fmt.Println("serving...")
	ticker := time.NewTicker(maxTimeout)
	defer ticker.Stop()
	select {
	case finishFunc := <-distributionCh:
		defer finishFunc()
		f.Delegate(resp, req)
	case <-ticker.C:
		fmt.Println("timed out...")
		resp.WriteHeader(409)
	}

}

const (
	maxTimeout = time.Minute * 10
)

func NewFQFilter(queues []*Queue) *FQFilter {

	// Initializing everything
	quotaCh := make(chan interface{})
	drainer := newQueueDrainer(queues, quotaCh)
	inflightFilter := &FQFilter{
		queues: queues,

		queueDrainer: drainer,

		sharedQuotaManager: newSharedQuotaManager(quotaCh, queues),

		Delegate: func(resp http.ResponseWriter, req *http.Request) {
			time.Sleep(time.Millisecond * 100) // assuming that it takes 100ms to finish the request
			resp.Write([]byte("success!"))
		},
	}

	return inflightFilter
}

func (f *FQFilter) Run() {
	go f.sharedQuotaManager.Run()
	go f.queueDrainer.Run()
}
