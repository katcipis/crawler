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

	const timeout = 30 * time.Second
	const reqTimeout = 5 * time.Second
	const concurrency = 30

	sites := map[string]uint{
		"https://google.com": 200,
		"https://bing.com":   1000,
		"https://monzo.com":  15000,
	}

	for site, minLinks := range sites {
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

		linksCount := uint(0)
		for res := range results {
			checkDomain(res.Parent)
			checkDomain(res.Link)
			checkRepeated(res)
			linksCount += 1
		}

		if linksCount < minLinks {
			t.Fatalf("expected at least [%d] links from [%s] and got[%d]", minLinks, site, linksCount)
		}
	}
}
