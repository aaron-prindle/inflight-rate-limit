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

func (f *FQFilter) Serve(resp http.ResponseWriter, req *http.Request) {

	// 0. Matching request w/ bindings API
	matchedQueue := findMatchedQueue(req, f.queues)

	// 1. Waiting to be notified by either a reserved quota or a shared quota
	distributionCh := f.queueDrainer.Enqueue(matchedQueue)
	if distributionCh == nil {
		// too many requests
		fmt.Println("throttled...")
		resp.WriteHeader(http.StatusConflict)
	}

	// fmt.Println("serving...")
	select {
	case finishFunc := <-distributionCh:
		defer finishFunc()
		f.Delegate(resp, req)
	// correctly handles the timeout handlers context
	case <-req.Context().Done():
		fmt.Println("timed out...")
		resp.WriteHeader(http.StatusConflict)
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
			resp.Write([]byte("success!\n"))
		},
	}

	return inflightFilter
}

func (f *FQFilter) Run() {
	go f.sharedQuotaManager.Run()
	go f.queueDrainer.Run()
}

func findMatchedQueue(req *http.Request, queues []*Queue) *Queue {
	priority := req.Header.Get("PRIORITY")
	idx, err := strconv.Atoi(priority)
	if err != nil {
		panic("strconv.Atoi(priority) errored")
	}
	return queues[idx]
}
