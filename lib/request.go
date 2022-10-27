package lib

import (
	"context"
	"encoding/base64"
	"github.com/levigross/grequests"
	"github.com/miekg/dns"
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

func ParseDNSMsg(dnsQuery string) dns.Msg {
	msg := dns.Msg{}
	pack, _ := base64.RawURLEncoding.DecodeString(dnsQuery)
	_ = msg.Unpack(pack)
	return msg
}

func ParseDomain(dnsQuery string) string {
	msg := ParseDNSMsg(dnsQuery)
	domain := msg.Question[0].Name
	return domain
}

func Resolve(host string, base string) (string, error) {
	msg := dns.Msg{}
	msg.SetQuestion(host+".", 1)
	buf, _ := msg.Pack()
	dnsQuery := base64.RawURLEncoding.EncodeToString(buf)
	uri := "https://" + base + "/dns-query?dns=" + dnsQuery
	resp, err := grequests.Get(uri, &grequests.RequestOptions{})
	defer func() {
		_ = resp.Close()
	}()
	if err != nil {
		return "", err
	}
	respMsg := dns.Msg{}
	_ = respMsg.Unpack(resp.Bytes())
	for _, item := range respMsg.Answer {
		if item.Header().Rrtype == 1 {
			return dns.Field(item, 1), nil
		}
	}
	return "", nil
}

type TestResult struct {
	Success bool   `json:"success"`
	Time    int64  `json:"time"`
	IP      string `json:"ip"`
}

func Test(host string, base string) TestResult {
	result := TestResult{
		Success: false,
		Time:    0,
		IP:      "",
	}
	if host != "" && base != "" {
		now := time.Now().UnixNano()
		ip, err := Resolve(host, base)
		result.IP = ip
		result.Success = err == nil
		result.Time = (time.Now().UnixNano() - now) / 1e6
	}
	return result
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
	ip, _ := Resolve(hostname, "225.5.5.5")
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
