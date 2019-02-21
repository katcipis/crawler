// Package crawler provides a concurrent crawler implementation
package crawler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
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
//
// All channels will be closed by the crawler when there is no more
// URLs to crawl or the provided context expires.
func Start(
	ctx context.Context,
	entrypoint url.URL,
	concurrency uint,
	timeout time.Duration,
) (<-chan Result, <-chan error) {

	res := make(chan Result, concurrency)
	errs := make(chan error, concurrency)

	if concurrency == 0 {
		go func() {
			errs <- errors.New("concurrency level must be greater than zero")
			close(errs)
			close(res)
		}()
		return res, errs
	}

	go scheduler(ctx, res, errs, entrypoint, concurrency, timeout)

	return res, errs
}

func scheduler(
	ctx context.Context,
	filtered chan<- Result,
	errs chan<- error,
	entrypoint url.URL,
	concurrency uint,
	timeout time.Duration,
) {
	defer close(filtered)
	defer close(errs)

	crawlResults := make(chan []Result, concurrency)
	defer close(crawlResults)

	jobs := make(chan url.URL, concurrency)
	defer close(jobs)

	for i := uint(0); i < concurrency; i++ {
		go crawler(ctx, jobs, timeout, crawlResults, errs)
	}

	pendingURLs := []url.URL{entrypoint}
	pendingJobs := 0
	filterByUniqueness := newUniquenessFilter(entrypoint)
	filterResByUniqueness := newResUniquenessFilter()

	for len(pendingURLs) > 0 || pendingJobs > 0 {

		jobsToSend := pendingURLs
		pendingJobs += len(jobsToSend)

		go func() {
			// WHY: create goroutine to avoid deadlock between
			// sending jobs to crawlers
			// and crawlers sending results back.
			for _, job := range jobsToSend {
				jobs <- job
			}
		}()

		results := filterBySameDomain(<-crawlResults)
		results = filterSelfReferences(results)
		results = filterResByUniqueness(results)
		pendingJobs -= 1

		for _, res := range results {
			filtered <- res
		}

		pendingURLs = filterByUniqueness(extractLinks(results))
	}
}

// crawler will write one set (possibly empty) of results for each
// job it reads from the jobs channel. Even on errors a empty results will
// be written, so the caller can trust that after writing N jobs it can
// expect N results. The errs channel will be used a side band of informational
// errors about the crawling process and should be drained.
func crawler(
	ctx context.Context,
	jobs <-chan url.URL,
	timeout time.Duration,
	res chan<- []Result,
	errs chan<- error,
) {
	client := &http.Client{Timeout: timeout}

	for url := range jobs {
		nextLinks, err := getLinks(ctx, client, url)

		if err != nil {
			errs <- err
			res <- nil
			continue
		}

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

func getLinks(ctx context.Context, c *http.Client, u url.URL) ([]url.URL, error) {
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("unable create GET request for url[%s]: %s", u.String(), err)
	}

	req = req.WithContext(ctx)
	res, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to GET url[%s]: %s", u.String(), err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"error status code[%d] on GET url[%s]",
			res.StatusCode,
			u.String())
	}

	//WHY: The web is a fierce jungle, it seems better to not trust
	//     HTTP headers and just try to parse the body searching for links
	links, err := parser.ExtractLinks(res.Body)
	if err != nil {
		return nil, fmt.Errorf(
			"error parsing response body from GET url[%s]: %s",
			u.String(),
			err)
	}

	absLinks := make([]url.URL, len(links))

	for i, link := range links {
		absLinks[i] = makeLinkAbsolute(u, link)
	}

	return absLinks, nil
}

func makeLinkAbsolute(parent url.URL, link url.URL) url.URL {
	if link.Host == "" {
		link.Host = parent.Host
	}
	if link.Scheme == "" {
		link.Scheme = parent.Scheme
	}
	if link.Path == "" {
		return link
	}
	if link.Path[0] != '/' {
		link.Path = path.Clean(parent.Path + "/" + link.Path)
	}
	if link.Path == "/" {
		link.Path = ""
	}
	return link
}

func newResUniquenessFilter() func([]Result) []Result {
	seen := map[string]bool{}

	return func(results []Result) []Result {
		filtered := []Result{}
		for _, res := range results {
			restr := res.String()
			if !seen[restr] {
				filtered = append(filtered, res)
				seen[restr] = true
			}

		}
		return filtered
	}
}

func newUniquenessFilter(entrypoint url.URL) func([]url.URL) []url.URL {
	seen := map[string]bool{
		entrypoint.String(): true,
	}

	return func(urls []url.URL) []url.URL {
		filtered := []url.URL{}
		for _, u := range urls {
			ustr := u.String()
			if !seen[ustr] {
				filtered = append(filtered, u)
				seen[ustr] = true
			}

		}
		return filtered
	}
}

func filterSelfReferences(results []Result) []Result {
	filtered := []Result{}

	for _, res := range results {
		if res.Link.String() != res.Parent.String() {
			filtered = append(filtered, res)
		}
	}

	return filtered
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
