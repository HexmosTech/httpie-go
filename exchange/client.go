package exchange

import (
	"crypto/tls"
	"net/http"
)

func BuildHTTPClient(options *Options,proxyURL string, autoRedirect bool) (*http.Client, error) {

	if autoRedirect {
        checkRedirect = nil // Follow redirects
    } else {
	checkRedirect := func(req *http.Request, via []*http.Request) error {
		// Do not follow redirects
		return http.ErrUseLastResponse
	}
	}
	// if options.FollowRedirects {
	// 	checkRedirect = nil
	// }

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
			proxyURLParsed, err := url.Parse(proxyURL)
			if err != nil {
				return nil, err
			}
			httpTransport.Proxy = http.ProxyURL(proxyURLParsed)
		}

		httpTransport.TLSClientConfig.InsecureSkipVerify = options.SkipVerify
		if options.ForceHTTP1 {
			httpTransport.TLSClientConfig.NextProtos = []string{"http/1.1", "http/1.0"}
			httpTransport.TLSNextProto = make(map[string]func(string, *tls.Conn) http.RoundTripper)
		}
	}
	client.Transport = transp

	return &client, nil
}
