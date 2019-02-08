// Package crawler provides a concurrent crawler implementation
package crawler

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/katcipis/crawler/parser"
)

// Result represents a single result found during the crawling process
type Result struct {
	// Link is the reachable link URL
	Link url.URL
	// Parent is the URL used to reach the Link URL
	Parent url.URL
}

func (r Result) String() string {
	return fmt.Sprintf("%s->%s", r.Parent.String(), r.Link.String())
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
// Both the result and the errors channels must be drained.
// If the caller reads only from the results channel the crawlers
// may become blocked writing errors.
func Start(
	entrypoint url.URL,
	concurrency uint,
	timeout time.Duration,
) (<-chan Result, <-chan error) {

	res := make(chan Result)
	errs := make(chan error)

	if concurrency == 0 {
		go func() {
			errs <- errors.New("concurrency level must be greater than zero")
			close(errs)
			close(res)
		}()
		return res, errs
	}

	go scheduler(res, errs, entrypoint, concurrency, timeout)

	return res, errs
}

func scheduler(
	filtered chan<- Result,
	errs chan<- error,
	entrypoint url.URL,
	concurrency uint,
	timeout time.Duration,
) {
	defer close(filtered)
	defer close(errs)

	crawlResults := make(chan []Result)
	defer close(crawlResults)

	jobs := make(chan url.URL)
	defer close(jobs)

	for i := uint(0); i < concurrency; i++ {
		go crawler(crawlResults, errs, jobs, timeout)
	}

	pendingURLs := []url.URL{entrypoint}
	filterByUniqueness := newUniquenessFilter()

	for len(pendingURLs) > 0 {
		jobsToSend := pendingURLs
		jobsSent := len(jobsToSend)

		go func() {
			// WHY: create goroutine to avoid deadlock between
			// sending jobs to crawlers
			// and crawlers sending results back.
			for _, job := range jobsToSend {
				jobs <- job
			}
		}()

		pendingURLs = nil
		for i := 0; i < jobsSent; i++ {
			results := filterBySameDomain(<-crawlResults)

			for _, res := range results {
				filtered <- res
			}

			pendingURLs = append(pendingURLs,
				filterByUniqueness(extractLinks(results))...)
		}
	}
}

func crawler(
	res chan<- []Result,
	errs chan<- error,
	jobs <-chan url.URL,
	timeout time.Duration,
) {
	client := &http.Client{Timeout: timeout}

	for url := range jobs {
		nextLinks := getLinks(client, url)
		results := make([]Result, len(nextLinks))

		for i, link := range nextLinks {
			results[i] = Result{
				Parent: url,
				Link:   link,
			}
		}

		res <- results
	}
}

func getLinks(c *http.Client, u url.URL) []url.URL {
	// TODO: improve error handling
	res, err := c.Get(u.String())
	if err != nil {
		return nil
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil
	}

	links, _ := parser.ExtractLinks(res.Body)
	absLinks := make([]url.URL, len(links))

	for i, link := range links {
		if link.Host == "" {
			link.Host = u.Host
		}
		if link.Scheme == "" {
			link.Scheme = u.Scheme
		}
		if !strings.HasPrefix(link.Path, "/") {
			link.Path = u.Path + "/" + link.Path
		}
		absLinks[i] = link
	}

	return absLinks
}

func newUniquenessFilter() func([]url.URL) []url.URL {
	seen := map[string]bool{}

	return func(urls []url.URL) []url.URL {
		filtered := []url.URL{}
		for _, u := range urls {
			ustr := u.String()

			if seen[ustr] {
				continue
			}

			filtered = append(filtered, u)
			seen[ustr] = true
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
