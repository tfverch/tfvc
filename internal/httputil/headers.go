package httputil

import (
	"fmt"
	"net/http"
)

type AddHeadersRoundtripper struct {
	Headers http.Header
	Nested  http.RoundTripper
}

func (h AddHeadersRoundtripper) RoundTrip(r *http.Request) (*http.Response, error) {
	for k, vs := range h.Headers {
		for _, v := range vs {
			r.Header.Add(k, v)
		}
	}
	resp, err := h.Nested.RoundTrip(r)
	if err != nil {
		return nil, fmt.Errorf("error http round trip %w", err)
	}
	return resp, nil
}
