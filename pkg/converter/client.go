package converter

import (
	"fmt"
	"io"
	"net/http"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type RealClient struct{}

func (RealClient) Do(req *http.Request) (*http.Response, error) {
	return performRequest(req)
}

// performRequest creates default HTTP Client and performs the request [req]. Returns error if it is unsuccessful,
// the response object otherwise.
func performRequest(req *http.Request) (*http.Response, error) {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request %s: %w", req.URL.String(), err)
	}

	return resp, nil
}

// checkRequestStatus checks status code of the response
func checkResponseStatus(resp *http.Response, url string) error {
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("api request failed. URL: <%s>, Code: %d, response: %s",
			url, resp.StatusCode, string(body))
	}

	return nil
}
