package service

import (
	"doh-go/lib"
	"os"
	"path"
)

var Store = lib.Store{}

func init() {
	cwd, _ := os.Getwd()
	storeDir := path.Join(cwd, "store")
	_ = os.Mkdir(storeDir, os.ModePerm)
	file := path.Join(storeDir, "data.json")
	Store.Init(file)
}
