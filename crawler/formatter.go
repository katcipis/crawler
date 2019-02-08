package crawler

import (
	"fmt"
	"io"
)

// Formatter is a function that given a channel of crawling results
// will write the results to the given writer according to a specific format
//
// The formatter will only return after draining the results channel
// and having written all the formatted data.
//
// If there is an error writing the formatted results an error is returned,
// otherwise the error will be nil.
type Formatter func(<-chan Result, io.Writer) error

// FormatAsTextSitemap will drain the given Result channel and
// write then in the given writer formatted as a text sitemap
// following this specification:
//
// https://www.sitemaps.org/protocol.html
//
// No repeated URLs are going to be written on the sitemap.
// The space complexity of this function is linear ( O(N) ) to
// the amount of unique URLs found in the results.
func FormatAsTextSitemap(res <-chan Result, w io.Writer) error {
	seen := map[string]bool{}
	first := true

	write := func(s string) error {
		if !seen[s] {
			seen[s] = true
			if first {
				first = false
			} else {
				s = "\n" + s
			}
			_, err := w.Write([]byte(s))
			return err
		}
		return nil
	}

	for r := range res {
		err := write(r.Parent.String())
		if err != nil {
			return fmt.Errorf("text sitemap formatter: failed to write result: %s", err)
		}

		err = write(r.Link.String())
		if err != nil {
			return fmt.Errorf("text sitemap formatter: failed to write result: %s", err)
		}
	}

	return nil
}
