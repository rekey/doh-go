package lib

import (
	"context"
	"encoding/json"
	"github.com/levigross/grequests"
	"net"
	"net/http"
	"net/url"
	"time"
)

type DNSAnswer struct {
	Name string `json:"name"`
	Data string `json:"data"`
	Type int    `json:"type"`
}

func Resolve(host string) string {
	uri := "https://223.5.5.5/resolve?name=" + host
	resp, _ := grequests.Get(uri, &grequests.RequestOptions{})
	defer func() {
		_ = resp.Close()
	}()
	result := struct {
		Answer []DNSAnswer
	}{}
	_ = json.Unmarshal(resp.Bytes(), &result)
	for _, item := range result.Answer {
		if item.Type == 1 {
			return item.Data
		}
	}
	return ""
}

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
			ctx context.Context, network, addr string,
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

func Request(uri string) (*grequests.Response, error) {
	u, _ := url.Parse(uri)
	hostname := u.Hostname()
	ip := Resolve(hostname)
	return grequests.Get(uri, &grequests.RequestOptions{
		HTTPClient: getHttpClient(hostname, ip, u.Port()),
	})
}
