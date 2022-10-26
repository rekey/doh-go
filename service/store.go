package service

import (
	"doh-go/lib"
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
		resp, _ := lib.Request("https://cdn.jsdelivr.net/gh/rekey/doh-go@main/store/data.json")
		err = os.WriteFile(file, resp.Bytes(), os.ModePerm)
		log.Println(err)
	}
	Store.Init(file)
}
