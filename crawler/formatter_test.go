package crawler_test

import (
	"bytes"
	"errors"
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

func TestOnWriteErrorFormatterFails(t *testing.T) {
	type tcase struct {
		name   string
		format crawler.Formatter
	}

	cases := []tcase{
		{
			name:   "TextSitemap",
			format: crawler.FormatAsTextSitemap,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res := make(chan crawler.Result)
			resCount := 2
			go func() {
				for i := 0; i < resCount; i++ {
					res <- crawler.Result{
						Parent: url.URL{
							Scheme: "http",
							Host:   "fail.com",
						},
					}
				}
				close(res)
			}()
			err := c.format(res, &explodingWriter{failOnCall: 1})
			if err == nil {
				t.Fatal("expected error on failed first write")
			}

			err = c.format(res, &explodingWriter{failOnCall: 2})
			if err == nil {
				t.Fatal("expected error on failed second write")
			}
		})
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

type explodingWriter struct {
	call       int
	failOnCall int
}

func (w *explodingWriter) Write(d []byte) (int, error) {
	w.call += 1
	if w.call == w.failOnCall {
		return 0, errors.New("exploding writer exploding !!")
	}
	return len(d), nil
}
