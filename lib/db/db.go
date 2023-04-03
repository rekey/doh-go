package db

import (
	"os"
	"path"

	"doh-go/lib/q"
)

type DB struct {
	Dir string
	DNS string
}

func (that *DB) Get(key string) string {
	filename := key + ".json"
	file := path.Join(that.Dir, filename)
	_, err := os.Stat(file)
	if err != nil && os.IsNotExist(err) {
		resp, _ := q.RequestIP("https://cdn.jsdelivr.net/gh/rekey/doh-go@main/store/"+filename, that.DNS)
		_ = os.WriteFile(file, resp.Bytes(), os.ModePerm)
		return string(resp.Bytes())
	}
	buf, _ := os.ReadFile(file)
	return string(buf)
}

func (that *DB) GetFile(file string) string {
	file = path.Join(that.Dir, file)
	_, err := os.Stat(file)
	if err != nil {
		return ""
	}
	buf, _ := os.ReadFile(file)
	return string(buf)
}

func (that *DB) Set(key string, data []byte) error {
	filename := key + ".json"
	file := path.Join(that.Dir, filename)
	return os.WriteFile(file, data, os.ModePerm)
}
