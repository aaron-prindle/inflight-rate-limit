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

	*sharedQuotaManager
	*queueDrainer

	Matcher  func(*http.Request, []*Queue) *Queue
	Delegate http.HandlerFunc
}

func (f *FQFilter) Serve(resp http.ResponseWriter, req *http.Request) {

	// 0. Matching request w/ bindings API
	matchedQueue := f.Matcher(req, f.queues)

	// 1. Waiting to be notified by either a reserved quota or a shared quota
	fmt.Println("1")
	distributionCh := f.queueDrainer.Enqueue(matchedQueue)
	if distributionCh == nil {
		// too many requests
		fmt.Println("throttled...")
		resp.WriteHeader(http.StatusConflict)
	}
	fmt.Println("2")
	// fmt.Println("serving...")
	select {
	case finishFunc := <-distributionCh:
		defer finishFunc()
		f.Delegate(resp, req)
	// supports the timeout handlers context
	case <-req.Context().Done():
		resp.WriteHeader(http.StatusConflict)
	}
}

func NewFQFilter(queues []*Queue) *FQFilter {

	// Initializing everything
	quotaCh := make(chan interface{})
	drainer := newQueueDrainer(queues, quotaCh)
	inflightFilter := &FQFilter{
		queues:             queues,
		queueDrainer:       drainer,
		sharedQuotaManager: newSharedQuotaManager(quotaCh, queues, drainer),
		Matcher:            findMatchedQueue,
		Delegate: func(resp http.ResponseWriter, req *http.Request) {
			time.Sleep(time.Millisecond * 100) // assuming that it takes 100ms to finish the request
			resp.Write([]byte("success!\n"))
		},
	}

	return inflightFilter
}

func (f *FQFilter) Run() {
	go f.sharedQuotaManager.Run()
}

func findMatchedQueue(req *http.Request, queues []*Queue) *Queue {
	priority := req.Header.Get("PRIORITY")
	idx, err := strconv.Atoi(priority)
	if err != nil {
		panic("strconv.Atoi(priority) errored")
	}
	return queues[idx]
}
