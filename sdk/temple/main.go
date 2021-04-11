package temple

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"github.com/zooyer/jsons"
)

func Init(path string) error {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}
	var c = make(map[string]map[string]*cacheInfo)
	for _, file := range files {
		if !file.IsDir() {
			continue
		}
		ns, err := readNamespace(filepath.Join(path, file.Name()))
		if err != nil {
			return err
		}
		c[file.Name()] = ns
	}
	cache = c
	dir = path
	return nil
}

func Notify(namespace, name string, fn func()) (err error) {
	return notify(namespace, name, fn)
}

func GetData(namespace, name string) (data []byte, err error) {
	return getData(namespace, name)
}

func GetConfig(namespace, name string, v interface{}) (err error) {
	data, err := getData(namespace, name)
	if err != nil {
		return
	}

	var raw []jsons.Raw
	if err = json.Unmarshal(data, &raw); err != nil {
		return
	}

	for i, r := range raw {
		raw[i] = r.Get("config")
	}

	if data, err = json.Marshal(raw); err != nil {
		return
	}

	return json.Unmarshal(data, v)
}
