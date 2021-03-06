// Package parser provides functions to parse HTML content
package parser

import (
	"fmt"
	"io"
	"net/url"

	"golang.org/x/net/html"
)

// ExtractLinks will return a list of links extracted
// from the given HTML body. If the content is not valid
// HTML an error is returned instead.
func ExtractLinks(htmlbody io.Reader) ([]url.URL, error) {
	doc, err := html.Parse(htmlbody)

	if err != nil {
		return nil, fmt.Errorf("parser.ExtractLinks: %s", err)
	}

	urls := []url.URL{}

	var visit func(n *html.Node)

	visit = func(n *html.Node) {
		if isLink(n) {
			extractedURL, ok := extractURL(n)
			if ok {
				urls = append(urls, extractedURL)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			visit(c)
		}
	}

	visit(doc)

	return urls, nil
}

func extractURL(linkNode *html.Node) (url.URL, bool) {
	for _, attr := range linkNode.Attr {
		if attr.Key == "href" {
			u, err := url.Parse(attr.Val)
			if err != nil {
				return url.URL{}, false
			}
			// WHY: Got some empty URLs being parsed on wikipedia.org
			if u.String() == "" {
				return url.URL{}, false
			}
			return *u, true
		}
	}

	return url.URL{}, false
}

func isLink(n *html.Node) bool {
	return n.Type == html.ElementNode && n.Data == "a"
}
