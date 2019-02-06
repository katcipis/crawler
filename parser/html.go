// Package parser provides functions to parse HTML content
package parser

import (
	"io"
	"net/url"

	"golang.org/x/net/html"
)

// ExtractLinks will return a list of links extracted
// from the given HTML body. If the content is not valid
// HTML an error is returned instead.
func ExtractLinks(htmlbody io.Reader) ([]url.URL, error) {
	// TODO: handle err
	doc, _ := html.Parse(htmlbody)
	urls := []url.URL{}

	var visit func(n *html.Node)

	visit = func(n *html.Node) {
		if isLink(n) {
			extractedURL := extractURL(n)
			urls = append(urls, extractedURL)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			visit(c)
		}
	}

	visit(doc)

	return urls, nil
}

func extractURL(linkNode *html.Node) url.URL {
	for _, attr := range linkNode.Attr {
		if attr.Key == "href" {
			// TODO: error handling
			u, _ := url.Parse(attr.Val)
			return *u
		}
	}

	// TODO: error handling
	return url.URL{}
}

func isLink(n *html.Node) bool {
	return n.Type == html.ElementNode && n.Data == "a"
}
