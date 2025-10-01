package converter

import "net/http"

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type RealClient struct{}

func (RealClient) Do(req *http.Request) (*http.Response, error) {
	return performRequest(req)
}
