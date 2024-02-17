package exchange

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"fmt"
)

func BuildHTTPClient(options *Options, proxyURL string,proxyUsername string, proxyPassword string, autoRedirect bool) (*http.Client, error) {
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

	var transp http.RoundTripper
	if options.Transport == nil {
		transp = http.DefaultTransport.(*http.Transport).Clone()
	} else {
		transp = options.Transport
	}
	if httpTransport, ok := transp.(*http.Transport); ok {
		if proxyURL != "" {
			fmt.Println("Inside HTTPclient setup , proxy assignment")
			proxyURLParsed, err := url.Parse(proxyURL)
			if err != nil {
				fmt.Println("Error parsing proxy URL:", err)
				return nil, err
			}

			proxyURLParsed.User = url.UserPassword(proxyUsername, proxyPassword)
			httpTransport.Proxy = http.ProxyURL(proxyURLParsed)
			

		}

		httpTransport.TLSClientConfig.InsecureSkipVerify = options.SkipVerify
		if options.ForceHTTP1 {
			httpTransport.TLSClientConfig.NextProtos = []string{"http/1.1", "http/1.0"}
			httpTransport.TLSNextProto = make(map[string]func(string, *tls.Conn) http.RoundTripper)
		}

		httpTransport.TLSClientConfig.InsecureSkipVerify = true
        httpTransport.DisableKeepAlives = true

	}


	client.Transport = transp
	fmt.Println("Configured http.Client:", client)
	return &client, nil
}
