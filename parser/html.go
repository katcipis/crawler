// Package parser provides functions to parse HTML content
package parser

import (
	"io"
	"net/url"
)

// ExtractURLs will return a list of URLs extracted
// from the given HTML body. If the content is not valid
// HTML an error is returned.
func ExtractURLs(htmlbody io.Reader) ([]url.URL, error) {
	return nil, nil
}
