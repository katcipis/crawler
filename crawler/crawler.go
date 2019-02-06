// Package crawler provides a concurrent crawler implementation
package crawler

import (
	"net/url"
	"time"
)

// Result represents a single result found during the crawling process
type Result struct {
	// Link is the reachable link URL
	Link url.URL
	// Parent is the URL used to reach the Link URL
	Parent url.URL
}

// Start will start N concurrent crawlers and return a channel
// where all results from the crawling can be received.
//
// The concurrency parameter controls how much concurrent
// crawlers will be started and the timeout controls the
// timeout of each request made. It is an error to pass
// 0 as the concurrency parameter.
//
// The crawler will only follow links from the same domain
// of the provided entry point URL.
//
// If the given entry point URL does not exist or any parameter
// is invalid it will return a nil channel and a non nil error
// with further details.
func Start(
	entrypoint url.URL,
	concurrency uint,
	timeout time.Duration,
) (<-chan Result, error) {
	// TODO: validate concurrency > 0
	res := make(chan Result, concurrency)
	go scheduler(res, entrypoint, concurrency, timeout)
	return res, nil
}

func scheduler(
	filtered chan<- Result,
	entrypoint url.URL,
	concurrency uint,
	timeout time.Duration,
) {
	res := make(chan []Result, concurrency)
	jobs := make(chan url.URL, concurrency)

	defer close(res)
	defer close(filtered)
	defer close(jobs)

	for i := uint(0); i < concurrency; i++ {
		go crawler(res, jobs, timeout)
	}

	pendingURLs := []url.URL{entrypoint}
	filterByUniqueness := newUniquenessFilter()

	for len(pendingURLs) > 0 {
		// WHY: avoid deadlock between sending jobs to crawlers
		// and crawlers sending results back.
		jobsToSend := pendingURLs
		jobsSent := len(jobsToSend)

		go func() {
			for _, job := range jobsToSend {
				jobs <- job
			}
		}()

		pendingURLs = nil
		for i := 0; i < jobsSent; i++ {
			results := <-res
			results = filterByUniqueness(results)
			results = filterByDomain(results, entrypoint.Host)
			pendingURLs = append(pendingURLs, extractLinks(results)...)
		}
	}
}

func crawler(
	res chan<- []Result,
	jobs <-chan url.URL,
	timeout time.Duration,
) {
	for range jobs {
		// TODO: do crawling
		res <- nil
	}
}

func newUniquenessFilter() func([]Result) []Result {
	// TODO
	return func(results []Result) []Result {
		return results
	}
}

func filterByDomain(results []Result, domain string) []Result {
	// TODO
	return results
}

func extractLinks(results []Result) []url.URL {
	// TODO
	return nil
}
