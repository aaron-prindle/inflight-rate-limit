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

	*sharedDispatcher

	Matcher  func(*http.Request, []*Queue) *Queue
	Delegate http.HandlerFunc
}

func (f *FQFilter) Serve(resp http.ResponseWriter, req *http.Request) {

	matchedQueue := f.Matcher(req, f.queues)

	dispatcher := f.sharedDispatcher.producers[matchedQueue.Priority]
	func() {
		dispatcher.lock.Lock()
		defer dispatcher.lock.Unlock()
		if dispatcher.requestsexecuting > dispatcher.ACV {
			// too many requests
			fmt.Println("throttled...")
			resp.WriteHeader(http.StatusConflict)
			return
		}
	}()
	distributionCh := dispatcher.fqScheduler.Enqueue(matchedQueue)
	if distributionCh == nil {
		// queues are full
		fmt.Println("throttled...")
		resp.WriteHeader(http.StatusConflict)
		return
	}
	select {
	case finishFunc := <-distributionCh:
		func() {
			dispatcher.lock.Lock()
			defer dispatcher.lock.Unlock()
			dispatcher.requestsexecuting++
		}()
		defer finishFunc()

		f.Delegate(resp, req)
	// supports the timeout handlers context
	case <-req.Context().Done():
		resp.WriteHeader(http.StatusConflict)
	}
}

func NewFQFilter(queues []*Queue) *FQFilter {
	// Initializing everything
	inflightFilter := &FQFilter{
		queues:           queues,
		sharedDispatcher: newSharedDispatcher(queues),
		Matcher:          findMatchedQueue,
		Delegate: func(resp http.ResponseWriter, req *http.Request) {
			time.Sleep(time.Millisecond * 100) // assuming that it takes 100ms to finish the request
			resp.Write([]byte("success!\n"))
		},
	}

	return inflightFilter
}

func (f *FQFilter) Run() {
	go f.sharedDispatcher.Run()
}

func findMatchedQueue(req *http.Request, queues []*Queue) *Queue {
	priority := req.Header.Get("PRIORITY")
	idx, err := strconv.Atoi(priority)
	if err != nil {
		panic("strconv.Atoi(priority) errored")
	}
	return queues[idx]
}
