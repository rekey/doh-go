package lib

import (
	"encoding/json"
	"io"
	"log"
	"math/rand"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/forease/gotld"
	"github.com/levigross/grequests"
	"github.com/miekg/dns"
)

type domianData struct {
	China map[string]int `json:"china"`
	GFW   map[string]int `json:"gfw"`
}

type domainUpdate struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type domianDNS struct {
	Global []string `json:"global"`
	China  []string `json:"china"`
}

type data struct {
	DNS     domianDNS      `json:"dns"`
	Update  []domainUpdate `json:"update"`
	Domains domianData     `json:"domains"`
}

type Store struct {
	data   data
	Dir    string
	logger *log.Logger
}

func (that *Store) getDomains(cate string) map[string]int {
	data := that.data.Domains.China
	if cate == "gfw" {
		data = that.data.Domains.GFW
	}
	return data
}

func (that *Store) AddDomain(key string, cate string) {
	key = strings.TrimSpace(key)
	if key == "" {
		return
	}
	domains := that.getDomains(cate)
	slice := strings.Split(key, ".")
	if len(slice) == 2 {
		domains[key] = 1
		return
	}
	_, domain, _ := gotld.GetTld(key)
	domains[key] = 1
	domains[domain] = 1
}

func (that *Store) GetDNSList(domain string) (string, []string) {
	GFWDomains := that.data.Domains.GFW
	GFWDNS := that.data.DNS.Global
	CNDomains := that.data.Domains.China
	CNDNS := that.data.DNS.China
	that.log("GFWDomains", domain, GFWDomains[domain])
	// 优先检测gfw
	if GFWDomains[domain] == 1 {
		return "global", GFWDNS
	}
	if CNDomains[domain] == 1 {
		return "china", CNDNS
	}
	_, ext, err := gotld.GetTld(domain)
	if err != nil {
		that.log("err", domain, err)
		return "china", CNDNS
	}
	slice := strings.Split(ext, ".")
	if ext == "top" {
		return "global", GFWDNS
	}
	if ext == "cn" {
		return "china", CNDNS
	}
	if slice[len(slice)-1] == "cn" {
		return "china", CNDNS
	}
	return "china", CNDNS
}

func (that *Store) GetDNS(key string) (string, string) {
	_, key, _ = gotld.GetTld(key)
	cate, list := that.GetDNSList(key)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	random := r.Intn(len(list))
	return cate, list[random]
}

func (that *Store) SaveDomains() {
	file := path.Join(that.Dir, "domains.json")
	china := path.Join(that.Dir, "china.json")
	gfw := path.Join(that.Dir, "gfw.json")
	buf, _ := json.Marshal(that.data.Domains)
	chinaBuf, _ := json.Marshal(that.data.Domains.China)
	gfwBuf, _ := json.Marshal(that.data.Domains.GFW)
	_ = os.WriteFile(file, buf, os.ModePerm)
	_ = os.WriteFile(china, chinaBuf, os.ModePerm)
	_ = os.WriteFile(gfw, gfwBuf, os.ModePerm)
}

func (that *Store) Save() {
	that.SaveDomains()
}

func (that *Store) Update() {
	that.log("update", "start")
	for _, item := range that.data.Update {
		resp, err := Request(item.Url)
		if err != nil {
			that.log("update", item.Name, item.Url, "done", "err:", err)
			continue
		}
		str := resp.String()
		lines := strings.Split(str, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			slice := strings.Split(line, "/")
			if len(slice) < 3 {
				continue
			}
			that.AddDomain(slice[1], item.Name)
		}
		that.log("update", item.Name, item.Url, "done", "err:", err)
	}
	that.updateCustomList("china")
	that.updateCustomList("gfw")
	that.Save()
	that.log("update", "done")
}

func (that *Store) updateCustomList(name string) {
	file := path.Join(that.Dir, name+".conf")
	_, err := os.Stat(file)
	if err != nil && os.IsNotExist(err) {
		return
	}
	buf, _ := os.ReadFile(file)
	lines := strings.Split(string(buf), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		that.log("custom", name, line)
		that.AddDomain(line, name)
	}
}

func (that *Store) initDNS() {
	file := path.Join(that.Dir, "dns.json")
	_, err := os.Stat(file)
	if err != nil && os.IsNotExist(err) {
		resp, _ := Request("https://cdn.jsdelivr.net/gh/rekey/doh-go@main/store/dns.json")
		_ = os.WriteFile(file, resp.Bytes(), os.ModePerm)
	}
	buf, _ := os.ReadFile(file)
	_ = json.Unmarshal(buf, &that.data.DNS)
}

func (that *Store) initUpdate() {
	file := path.Join(that.Dir, "update.json")
	_, err := os.Stat(file)
	if err != nil && os.IsNotExist(err) {
		resp, _ := Request("https://cdn.jsdelivr.net/gh/rekey/doh-go@main/store/update.json")
		_ = os.WriteFile(file, resp.Bytes(), os.ModePerm)
	}
	buf, _ := os.ReadFile(file)
	_ = json.Unmarshal(buf, &that.data.Update)
}

func (that *Store) initDomains() {
	file := path.Join(that.Dir, "domains.json")
	_, err := os.Stat(file)
	if err != nil && os.IsNotExist(err) {
		resp, _ := Request("https://cdn.jsdelivr.net/gh/rekey/doh-go@main/store/domains.json")
		_ = os.WriteFile(file, resp.Bytes(), os.ModePerm)
	}
	buf, _ := os.ReadFile(file)
	_ = json.Unmarshal(buf, &that.data.Domains)
}

func (that *Store) log(v ...any) {
	v = append([]any{time.Now().Format("2006-01-02 15:04:05") + ":"}, v...)
	that.logger.Println(v...)
}

func (that *Store) Resolve(query string) (*grequests.Response, error) {
	domain := ParseDomain(query)
	arr := dns.SplitDomainName(domain)
	cate, host := that.GetDNS(strings.Join(arr, "."))
	url := "https://" + host + "/dns-query?dns=" + query
	now := time.Now().UnixNano()
	resp, err := grequests.Get(url, &grequests.RequestOptions{})
	ms := (time.Now().UnixNano() - now) / 1e6
	that.log("query", domain, cate, host, strconv.FormatInt(ms, 10)+"ms")
	return resp, err
}

func (that *Store) Check(domain string) string {
	arr := dns.SplitDomainName(domain)
	cate, host := that.GetDNS(strings.Join(arr, "."))
	return cate + "|" + host
}

func (that *Store) Init(w io.Writer) {
	_ = os.Mkdir(that.Dir, os.ModePerm)
	that.data = data{
		DNS: domianDNS{},
		Domains: domianData{
			China: map[string]int{},
			GFW:   map[string]int{},
		},
		Update: []domainUpdate{},
	}
	that.logger = log.New(w, "[Flamego] ", 0)
	that.initDNS()
	that.initUpdate()
	that.initDomains()
}
