//go:build wasm

package exchange

import (
	"fmt"
	"net/http"
)

func BuildHTTPClient(options *Options, proxyURL string, proxyUsername string, proxyPassword string, autoRedirect bool, request *http.Request) (*http.Client, error) {
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
