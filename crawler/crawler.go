// Package crawler provides a concurrent crawler implementation
package crawler

import (
	"net/http"
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
	defer close(filtered)

	res := make(chan []Result, concurrency)
	defer close(res)

	jobs := make(chan url.URL)
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
			results = filterBySameDomain(results)
			results = filterByUniqueness(results)

			for _, res := range results {
				filtered <- res
			}

			pendingURLs = append(pendingURLs, extractLinks(results)...)
		}
	}
}

func crawler(
	res chan<- []Result,
	jobs <-chan url.URL,
	timeout time.Duration,
) {
	client := http.Client{Timeout: timeout}

	for url := range jobs {
		// TODO: do crawling
		client.Get(url.String())
		res <- nil
	}
}

func newUniquenessFilter() func([]Result) []Result {
	knows := map[string]bool{}

	return func(results []Result) []Result {
		filtered := []Result{}
		for _, res := range results {
			urlpair := res.Parent.String() + res.Link.String()

			if knows[urlpair] {
				continue
			}

			filtered = append(filtered, res)
			knows[urlpair] = true
		}
		return filtered
	}
}

func filterBySameDomain(results []Result) []Result {
	filtered := []Result{}

	for _, res := range results {
		if res.Link.Host == res.Parent.Host {
			filtered = append(filtered, res)
		}
	}

	return filtered
}

func extractLinks(results []Result) []url.URL {
	links := make([]url.URL, len(results))
	for i, res := range results {
		links[i] = res.Link
	}
	return links
}
