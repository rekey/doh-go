package q

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/levigross/grequests"
)

func getHttpClient(host string, ip string, port string) *http.Client {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	var transport http.RoundTripper = &http.Transport{
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DialContext: func(
			ctx context.Context,
			network,
			addr string,
		) (net.Conn, error) {
			matchAddr := host + ":" + port
			if addr == matchAddr {
				addr = ip + ":" + port
			}
			return dialer.DialContext(ctx, network, addr)
		},
	}
	return &http.Client{
		Transport: transport,
	}
}

func RequestIP(uri string, dns string) (*grequests.Response, error) {
	u, _ := url.Parse(uri)
	hostname := u.Hostname()
	ip, _ := Resolve(hostname, dns)
	port := u.Port()
	if port == "" {
		port = "80"
		if u.Scheme == "https" {
			port = "443"
		}
	}

	return grequests.Get(uri, &grequests.RequestOptions{
		HTTPClient: getHttpClient(hostname, ip, port),
	})
}
