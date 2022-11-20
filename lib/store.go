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
	w      io.Writer
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
	if domain == "" {
		domains[key] = 1
		return
	}
	domains[domain] = 1
}

func (that *Store) GetDNSList(key string) (string, []string) {
	_, key, _ = gotld.GetTld(key)
	slice := strings.Split(key, ".")
	if slice[len(slice)-1] == "cn" {
		return "china", that.data.DNS.China
	}
	if key == "cn" {
		return "china", that.data.DNS.China
	}
	if key == "top" {
		return "global", that.data.DNS.Global
	}
	// 优先检测gfw，后续如果gfw和china表冲突，优先匹配gfw
	domains := that.data.Domains.GFW
	if domains[key] == 1 {
		return "global", that.data.DNS.Global
	}
	domains = that.data.Domains.China
	if domains[key] == 1 {
		return "china", that.data.DNS.China
	}
	return "global", that.data.DNS.Global
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
	buf, _ := json.Marshal(that.data.Domains)
	_ = os.WriteFile(file, buf, os.ModePerm)
}

func (that *Store) Save() {
	that.SaveDomains()
}

func (that *Store) Update() {
	that.log("update", "start")
	for _, item := range that.data.Update {
		resp, err := Request(item.Url)
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
	that.Save()
	that.updateCustomList("china")
	that.updateCustomList("gfw")
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
		log.Println("custom", name, line)
		that.AddDomain(line, name)
	}
}

func (that *Store) initDNS() {
	file := path.Join(that.Dir, "dns.json")
	_, err := os.Stat(file)
	if err != nil && os.IsNotExist(err) {
		resp, _ := Request("https://cdn.jsdelivr.net/gh/rekey/doh-go@main/store/dns.json")
		err = os.WriteFile(file, resp.Bytes(), os.ModePerm)
	}
	buf, _ := os.ReadFile(file)
	_ = json.Unmarshal(buf, &that.data.DNS)
}

func (that *Store) initUpdate() {
	file := path.Join(that.Dir, "update.json")
	_, err := os.Stat(file)
	if err != nil && os.IsNotExist(err) {
		resp, _ := Request("https://cdn.jsdelivr.net/gh/rekey/doh-go@main/store/update.json")
		err = os.WriteFile(file, resp.Bytes(), os.ModePerm)
	}
	buf, _ := os.ReadFile(file)
	_ = json.Unmarshal(buf, &that.data.Update)
}

func (that *Store) initDomains() {
	file := path.Join(that.Dir, "domains.json")
	_, err := os.Stat(file)
	if err != nil && os.IsNotExist(err) {
		resp, _ := Request("https://cdn.jsdelivr.net/gh/rekey/doh-go@main/store/domains.json")
		err = os.WriteFile(file, resp.Bytes(), os.ModePerm)
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
