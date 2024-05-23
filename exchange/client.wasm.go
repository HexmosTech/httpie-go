//go:build wasm

package exchange

import (
	"net/http"
)

func BuildHTTPClient(options *Options, autoRedirect bool) (*http.Client, error) {
	var checkRedirect func(req *http.Request, via []*http.Request) error

	if autoRedirect {
		checkRedirect = nil // Follow redirects
	} else {
		checkRedirect = func(req *http.Request, via []*http.Request) error {
			// Do not follow redirects
			return http.ErrUseLastResponse
		}
	}

	client := &http.Client{
		CheckRedirect: checkRedirect,
	}

	return client, nil
}
