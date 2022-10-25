package lib

import (
	"encoding/json"
	"github.com/forease/gotld"
	"github.com/levigross/grequests"
	"math/rand"
	"os"
	"strings"
	"time"
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
	data data
	file string
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

func (that *Store) GetDNSList(key string) []string {
	_, key, _ = gotld.GetTld(key)
	slice := strings.Split(key, ".")
	if slice[len(slice)-1] == "cn" {
		return that.data.DNS.China
	}
	if key == "cn" {
		return that.data.DNS.China
	}
	if key == "top" {
		return that.data.DNS.Global
	}
	domains := that.data.Domains.China
	if domains[key] != 1 {
		return that.data.DNS.Global
	}
	return that.data.DNS.China
}

func (that *Store) GetDNS(key string) string {
	_, key, _ = gotld.GetTld(key)
	list := that.GetDNSList(key)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	random := r.Intn(len(list))
	return list[random]
}

func (that *Store) Save() {
	buf, _ := json.Marshal(that.data)
	_ = os.WriteFile(that.file, buf, os.ModePerm)
}

func (that *Store) Update() {
	for _, item := range that.data.Update {
		resp, _ := grequests.Get(item.Url, &grequests.RequestOptions{})
		str := resp.String()
		lines := strings.Split(str, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			slice := strings.Split(line, "/")
			if len(slice) < 3 {
				continue
			}
			that.AddDomain(slice[1], item.Name)
		}
	}
	that.Save()
}

func (that *Store) Init(file string) {
	that.file = file
	buf, _ := os.ReadFile(that.file)
	err := json.Unmarshal(buf, &that.data)
	if err != nil {
		that.data = data{
			DNS: domianDNS{},
			Domains: domianData{
				China: map[string]int{},
				GFW:   map[string]int{},
			},
			Update: []domainUpdate{},
		}
	}
}
