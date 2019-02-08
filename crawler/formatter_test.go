package crawler_test

import (
	"bytes"
	"net/url"
	"testing"

	"github.com/katcipis/crawler/crawler"
)

type FormatterTestCase struct {
	name    string
	results []crawler.Result
	want    string
}

func TestTextSitemapFormatter(t *testing.T) {
	// Based on these specs: https://www.sitemaps.org/protocol.html
	cases := []FormatterTestCase{
		{
			name:    "empty",
			results: []crawler.Result{},
			want:    "",
		},
		{
			name: "two",
			results: []crawler.Result{
				{
					Parent: url.URL{
						Host:   "test",
						Scheme: "http",
					},
					Link: url.URL{
						Host:   "test",
						Path:   "/link",
						Scheme: "https",
					},
				},
				{
					Parent: url.URL{
						Host:   "test",
						Path:   "/link",
						Scheme: "https",
					},
					Link: url.URL{
						Host:   "test:8888",
						Scheme: "http",
					},
				},
			},
			want: "http://test\nhttps://test/link\nhttp://test:8888",
		},
	}

	for _, c := range cases {
		testFormatter(t, c, crawler.FormatAsTextSitemap)
	}
}

func testFormatter(t *testing.T, c FormatterTestCase, format crawler.Formatter) {
	t.Run(c.name, func(t *testing.T) {
		res := make(chan crawler.Result)
		buffer := &bytes.Buffer{}

		go func() {
			for _, r := range c.results {
				res <- r
			}
			close(res)
		}()

		err := format(res, buffer)
		if err != nil {
			t.Fatal(err)
		}

		got := buffer.String()
		if c.want != got {
			t.Fatalf("want:[%s] != got[%s]", c.want, got)
		}
	})
}
