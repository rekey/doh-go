package service

import (
	"doh-go/lib"
	"github.com/levigross/grequests"
	"log"
	"os"
	"path"
)

var Store = lib.Store{}

func init() {
	cwd, _ := os.Getwd()
	storeDir := path.Join(cwd, "store")
	_ = os.Mkdir(storeDir, os.ModePerm)
	file := path.Join(storeDir, "data.json")
	_, err := os.Stat(file)
	if err != nil && os.IsNotExist(err) {
		resp, _ := grequests.Get(
			"https://cdn.jsdelivr.net/gh/rekey/doh-go/store/data.json",
			&grequests.RequestOptions{},
		)
		err = os.WriteFile(file, resp.Bytes(), os.ModePerm)
		log.Println(err)
	}
	Store.Init(file)
}
