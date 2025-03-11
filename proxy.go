package main

import (
	"fmt"
	"net/http"
	"net/url"
)

// ProxyRoundTripper is an implementation of http.RoundTripper that supports
// setting a proxy server URL for genai clients. This type should be used with
// a custom http.Client that's passed to WithHTTPClient. For such clients,
// WithAPIKey doesn't apply so the key has to be explicitly set here.
type ProxyRoundTripper struct {
	APIKey   string
	ProxyURL string
}

/*
RoundTrip implements the http.RoundTripper interface with proxy support for Gemini API requests. It modifies
outgoing HTTP requests to include the API key and route them through a configured proxy URL if provided,
handling internet proxy connections for Gemini API calls.
*/
func (t *ProxyRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()

	if t.ProxyURL != "" {
		proxyURL, err := url.Parse(t.ProxyURL)
		if err != nil {
			fmt.Printf("error [%v] parsing internet proxy url\n", err)
			return nil, err
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	newReq := req.Clone(req.Context())
	query := newReq.URL.Query()
	query.Add("key", t.APIKey)
	newReq.URL.RawQuery = query.Encode()

	resp, err := transport.RoundTrip(newReq)
	if err != nil {
		fmt.Printf("error [%v] setting up 'transport round trip' for internet proxy\n", err)
		return nil, err
	}

	return resp, nil
}
