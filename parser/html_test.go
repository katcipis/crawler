package parser_test

import (
	"bytes"
	"net/url"
	"testing"

	"github.com/katcipis/crawler/parser"
)

func TestExtractLinks(t *testing.T) {
	type tcase struct {
		name string
		html string
		want []string
	}

	cases := []tcase{
		{
			name: "oneLink",
			html: `<a href="/test"></a>`,
			want: []string{"/test"},
		},
		{
			name: "multipleLinksOnRoot",
			html: `<a href="/test"></a>`,
			want: []string{"/test"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			htmlReader := bytes.NewBufferString(c.html)
			gotURLs, err := parser.ExtractLinks(htmlReader)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			got := urlsAsStrings(gotURLs)

			if len(c.want) != len(got) {
				t.Fatalf("want '%s' != got '%s'", c.want, got)
			}

			for i, wantURL := range c.want {
				gotURL := got[i]
				if wantURL != gotURL {
					t.Errorf("want[%s] != got[%s] at index[%d]", wantURL, gotURL, i)
				}
			}
		})
	}
}

func urlsAsStrings(urls []url.URL) []string {
	urlsStr := make([]string, len(urls))
	for i, url := range urls {
		urlsStr[i] = url.String()
	}
	return urlsStr
}
