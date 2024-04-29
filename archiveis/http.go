package archiveis

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

func doRequest(method string, url string, body io.ReadCloser, timeout time.Duration) (*http.Response, []byte, error) {
	req, err := newRequest(method, url, body)
	if err != nil {
		return nil, nil, err
	}

	if method != "" && method != "get" {
		req.Header.Set("content-type", "application/x-www-form-urlencoded")
	}

	client := newClient(timeout)
	resp, err := client.Do(req)
	if err != nil {
		return resp, nil, fmt.Errorf("executing request: %s", err)
	}
	if resp.StatusCode/100 != 2 && resp.StatusCode/100 != 3 {
		return resp, nil, fmt.Errorf("%v request to %v received unhappy response status-code=%v", method, url, resp.StatusCode)
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, nil, fmt.Errorf("reading response body: %s", err)
	}
	if err := resp.Body.Close(); err != nil {
		return resp, respBody, fmt.Errorf("closing response body: %s", err)
	}
	return resp, respBody, nil
}

func newRequest(method string, url string, body io.ReadCloser) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("creating %v request to %v: %s", method, url, err)
	}

	req.Host = HTTPHost

	hostname := strings.Split(BaseURL, "://")[1]
	req.Header.Set("Host", hostname)
	req.Header.Set("Origin", hostname)
	req.Header.Set("Authority", hostname)
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Referer", BaseURL+"/")

	return req, nil
}

func newClient(timeout time.Duration) *http.Client {
	c := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			Dial: (&net.Dialer{
				Timeout:   timeout,
				KeepAlive: timeout,
			}).Dial,
			TLSHandshakeTimeout:   timeout,
			ResponseHeaderTimeout: timeout,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
	return c
}
