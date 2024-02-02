package exchange

import (
	"crypto/tls"
	"net/http"
	"net/url"
)

func BuildHTTPClient(options *Options, proxyURL string,proxyUsername string, proxyPassword string, autoRedirect bool) (*http.Client, error) {
	var checkRedirect func(req *http.Request, via []*http.Request) error

	if autoRedirect {
		checkRedirect = nil // Follow redirects
	} else {
		checkRedirect = func(req *http.Request, via []*http.Request) error {
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
				fmt.Println(err)
				return nil, err
			}

			proxyURLParsed.User = url.UserPassword(proxyUsername, proxyPassword)
			httpTransport.Proxy = http.ProxyURL(proxyURLParsed)
			
			// initial
			// httpTransport.Proxy = http.ProxyURL(proxyURLParsed)
			// New
			// httpTransport.ProxyConnect = func(network, addr string) (netConn net.Conn, err error) {
			// 	proxyURLParsed.User = url.UserPassword(proxyUsername, proxyPassword)
			// 	return httpTransport.ProxyConnect(network, proxyURLParsed.Host)
			// }

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
