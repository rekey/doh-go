package dns

import (
	"bytes"
	"doh-go/lib/db"
	"doh-go/lib/q"
	"encoding/json"
	"math/rand"
	"strings"
	"time"

	"github.com/forease/gotld"
)

type KV map[string]string
type TData map[string]KV

type DNS struct {
	DB    *db.DB
	Data  TData
	First TData
}

func (that *DNS) Init() {
	that.Data = TData{}
	that.First = TData{}
	arr := []string{
		"update",
		"china",
		"gfw",
		"dns",
	}
	for _, v := range arr {
		s := that.DB.Get(v)
		kv := KV{}
		_ = json.Unmarshal(bytes.NewBufferString(s).Bytes(), &kv)
		that.Data[v] = kv
	}
	that.Update()
	that.UpdateFirst()
	go func() {
		for {
			time.Sleep(time.Hour * 12)
			that.Update()
			that.UpdateFirst()
		}
	}()
}

func (that *DNS) Update() {
	data := that.Data["update"]
	for _, k := range []string{"china", "gfw"} {
		u := data[k]
		resp, err := q.RequestIP(u, "223.5.5.5")
		defer func() {
			_ = resp.Close()
		}()
		if err != nil {
			continue
		}
		lines := strings.Split(string(resp.Bytes()), "\r\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			lineArr := strings.Split(line, "/")
			if len(lineArr) < 3 {
				continue
			}
			domain := lineArr[1]
			that.Data[k][domain] = "1"
		}
		buf, _ := json.Marshal(that.Data[k])
		that.DB.Set(k, buf)
	}
}

func (that *DNS) UpdateFirst() {
	for _, k := range []string{"china", "gfw"} {
		s := that.DB.GetFile(k + ".conf")
		if s == "" {
			continue
		}
		lines := strings.Split(s, "\r\n")
		that.First[k] = KV{}
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			that.First[k][line] = "1"
		}
	}
}

func (that *DNS) GetDNSType(domain string) string {
	for _, k := range []string{"china", "gfw"} {
		if match(domain, that.First[k]) {
			return k
		}
	}
	for _, k := range []string{"china", "gfw"} {
		if match(domain, that.Data[k]) {
			return k
		}
	}
	return "gfw"
}

func (that *DNS) GetDNS(domain string) (string, string) {
	length := len(domain)
	lastS := domain[length-1]
	if string(lastS) == "." {
		domain = string(domain[0 : length-1])
	}
	t := that.GetDNSType(domain)
	list := strings.Split(that.Data["dns"][t], ",")
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	random := r.Intn(len(list))
	return t, list[random]
}

func match(domain string, domains KV) bool {
	if domains[domain] == "1" {
		return true
	}
	_, root, err := gotld.GetTld(domain)
	if err != nil {
		return false
	}
	if domains[root] == "1" {
		return true
	}
	return false
}
