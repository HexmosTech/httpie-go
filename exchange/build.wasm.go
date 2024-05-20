//go:build wasm

package exchange

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/HexmosTech/httpie-go/input"
)

func BuildHTTPRequest(in *input.Input, options *Options) (*http.Request, error) {
	originalURL, err := buildURL(in)
	if err != nil {
		return nil, err
	}
	proxyURL := "https://proxyserver.hexmos.com/"
	uString := fmt.Sprintf("%s%s", proxyURL, originalURL.String())

	modifiedURL, err := url.Parse(uString)
	if err != nil {
		return nil, err
	}

	header, err := buildHTTPHeader(in)
	if err != nil {
		return nil, err
	}

	bodyTuple, err := buildHTTPBody(in)
	if err != nil {
		return nil, err
	}

	if header.Get("Content-Type") == "" && bodyTuple.contentType != "" {
		header.Set("Content-Type", bodyTuple.contentType)
	}
	if header.Get("User-Agent") == "" {
		header.Set("User-Agent", `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/39.0.2171.27 Safari/537.36`)
	}

	r := http.Request{
		Method:        string(in.Method),
		URL:           modifiedURL,
		Header:        header,
		Host:          header.Get("Host"),
		Body:          bodyTuple.body,
		GetBody:       bodyTuple.getBody,
		ContentLength: bodyTuple.contentLength,
	}

	if options.Auth.Enabled {
		r.SetBasicAuth(options.Auth.UserName, options.Auth.Password)
	}

	return &r, nil
}
