// +build integration

package crawler_test

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/katcipis/crawler/crawler"
)

func TestCrawlingWebsites(t *testing.T) {

	const timeout = 20 * time.Second
	const reqTimeout = 5 * time.Second
	const concurrency = 4

	sites := []string{
		"https://google.com",
		"https://bing.com",
		"https://monzo.com",
	}

	for _, site := range sites {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		entrypoint, err := url.Parse(site)
		fatalerr(t, err, fmt.Sprintf("parsing site: %s", site))

		results, errs := crawler.Start(ctx, *entrypoint, concurrency, reqTimeout)

		go func() {
			for err := range errs {
				fmt.Fprintln(os.Stderr, err)
			}
		}()

		checkDomain := func(link url.URL) {
			if link.Host != entrypoint.Host {
				fmt.Errorf(
					"invalid link with wrong domain[%+v], expected domain is[%s]",
					link.String(),
					entrypoint.Host,
				)
			}
		}

		seen := map[string]bool{}
		checkRepeated := func(res crawler.Result) {
			restr := res.String()
			if seen[restr] {
				t.Fatalf("duplicated result: [%s]", restr)
			}
			seen[restr] = true
		}

		linksCount := 0
		for res := range results {
			checkDomain(res.Parent)
			checkDomain(res.Link)
			checkRepeated(res)
			linksCount += 1
		}

		if linksCount == 0 {
			t.Fatalf("expected at least one link from [%s]", site)
		}
	}
}
