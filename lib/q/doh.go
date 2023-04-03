package q

import (
	"encoding/base64"
	"time"

	"github.com/levigross/grequests"
	"github.com/miekg/dns"
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
	query := base64.RawURLEncoding.EncodeToString(buf)
	uri := "https://" + base + "/dns-query?dns=" + query
	resp, err := grequests.Get(uri, &grequests.RequestOptions{
		RequestTimeout: 30 * time.Second,
		DialKeepAlive:  30 * time.Second,
	})
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

func ResolveOrigin(query string, base string) (*grequests.Response, error) {
	url := "https://" + base + "/dns-query?dns=" + query
	return grequests.Get(url, &grequests.RequestOptions{
		RequestTimeout: 30 * time.Second,
		DialKeepAlive:  60 * time.Second,
	})
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
