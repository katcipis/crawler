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

// Start will start a concurrent crawler and return a channel
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
	res := make(chan Result)
	close(res)
	return res, nil
}
