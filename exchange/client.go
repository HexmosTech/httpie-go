package exchange

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
)

func BuildHTTPClient(options *Options, proxyURL string, proxyUsername string, proxyPassword string, autoRedirect bool) (*http.Client, error) {
	fmt.Println("inside BuildHTTPClient Function")
	var checkRedirect func(req *http.Request, via []*http.Request) error

	if autoRedirect {
		checkRedirect = nil // Follow redirects
	} else {
		checkRedirect = func(req *http.Request, via []*http.Request) error {
			// Do not follow redirects
			return http.ErrUseLastResponse
		}
	}

	client := http.Client{
		CheckRedirect: checkRedirect,
		Timeout:       options.Timeout,
	}
	 
	if proxyURL != "" {
		fmt.Println("Inside HTTP client setup, proxy assignment")

		proxyURLParsed, err := url.Parse(proxyURL)
		if err != nil {
			fmt.Println("Error parsing proxy URL:", err)
			return nil, err
		}
		proxyURLParsed.User = url.UserPassword(proxyUsername, proxyPassword)
		proxyTransport := &http.Transport{
			Proxy:           http.ProxyURL(proxyURLParsed),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: options.SkipVerify},
			DisableKeepAlives: true,
		}

		if options.ForceHTTP1 {
			proxyTransport.TLSNextProto = make(map[string]func(string, *tls.Conn) http.RoundTripper)
			proxyTransport.TLSClientConfig.NextProtos = []string{"http/1.1", "http/1.0"}
		}

		client.Transport = proxyTransport
 		fmt.Println("Configured http.Client with proxy:", client)
		return &client, nil

	}

	fmt.Println("Configured http.Client without proxy:", client)
	return &client, nil
}