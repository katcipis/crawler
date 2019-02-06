package parser_test

import (
	"errors"
	"net/url"
	"strings"
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
			name: "emptyContent",
			html: "",
			want: []string{},
		},
		{
			name: "noLinks",
			html: `<body></body>`,
			want: []string{},
		},
		{
			name: "oneLink",
			html: `<a href="/test"></a>`,
			want: []string{"/test"},
		},
		{
			name: "multipleLinks",
			html: `
				<a href="/test1"></a>
				<a href="/test2"></a>
				<a href="/test3"></a>
			`,
			want: []string{"/test1", "/test2", "/test3"},
		},
		{
			name: "ignoresLinksWithoutHref",
			html: `
				<a href="/test1"></a>
				<a nothref="/test2"></a>
				<a href="/test3"></a>
			`,
			want: []string{"/test1", "/test3"},
		},
		{
			name: "ignoresLinksWithInvalidHref",
			html: `
				<a href=":/invalid"></a>
				<a href="/test1"></a>
				<a href=":/invalid"></a>
				<a href="/test3"></a>
				<a href=":/invalid"></a>
			`,
			want: []string{"/test1", "/test3"},
		},
		{
			name: "linkWithSchemeAndDomain",
			html: `
				<a href="http://example.com"></a>
			`,
			want: []string{"http://example.com"},
		},
		{
			name: "linkWithPort",
			html: `
				<a href="http://example.com:7777"></a>
			`,
			want: []string{"http://example.com:7777"},
		},
		{
			name: "multipleNestedLinks",
			html: `
				<body>
					<p>
						<a href="http://coding.is.fun/test1"></a>
					</p>
					<a href="https://coding.is.fun/test2"></a>
					<h1>
						<a href="ftp://coding.is.fun/test3"></a>
					</h1>
					<a href="http://coding.is.fun"></a>
				</body>
			`,
			want: []string{
				"http://coding.is.fun/test1",
				"https://coding.is.fun/test2",
				"ftp://coding.is.fun/test3",
				"http://coding.is.fun",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			htmlReader := strings.NewReader(c.html)
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

func TestExtractLinksFailsOnReadError(t *testing.T) {
	res, err := parser.ExtractLinks(&explodingReader{})
	if err == nil {
		t.Fatalf("expected error, instead got valid result: %v", res)
	}
}

type explodingReader struct{}

func (*explodingReader) Read([]byte) (int, error) {
	return 0, errors.New("explodingReader error")
}

func urlsAsStrings(urls []url.URL) []string {
	urlsStr := make([]string, len(urls))
	for i, url := range urls {
		urlsStr[i] = url.String()
	}
	return urlsStr
}
