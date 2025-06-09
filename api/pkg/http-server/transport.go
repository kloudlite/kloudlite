package httpServer

import "net/http"

type roundTripperWithHeaders struct {
	//read more at https://pkg.go.dev/net/http#Header
	headers map[string][]string
	rt      http.RoundTripper
}

func (rt roundTripperWithHeaders) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, vv := range rt.headers {
		for _, v := range vv {
			req.Header.Add(k, v)
		}
	}
	return rt.rt.RoundTrip(req)
}

var _ http.RoundTripper = (*roundTripperWithHeaders)(nil)

func NewRoundTripperWithHeaders(rt http.RoundTripper, headers map[string][]string) http.RoundTripper {
	return &roundTripperWithHeaders{
		headers: headers,
		rt:      rt,
	}
}
