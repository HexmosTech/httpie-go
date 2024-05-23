//go:build cli

package exchange

import (
	"fmt"
	"net/http"

	"github.com/HexmosTech/httpie-go/input"
	"github.com/HexmosTech/httpie-go/version"
)

func BuildHTTPRequest(in *input.Input, options *Options) (*http.Request, error) {
	originalURL, err := buildURL(in)
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
		header.Set("User-Agent", fmt.Sprintf("httpie-go/%s", version.Current()))
	}

	r := http.Request{
		Method:        string(in.Method),
		URL:           originalURL,
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
