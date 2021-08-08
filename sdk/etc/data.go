package etc

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type cacheInfo struct {
	Data      []byte
	Timestamp int64
}

var (
	dir   string
	mutex sync.RWMutex
	cache = make(map[string]map[string]*cacheInfo)
)

func readNamespace(dir string) (map[string]*cacheInfo, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var cache = make(map[string]*cacheInfo)
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		filename := filepath.Join(dir, file.Name())
		ft, err := fileTime(filename)
		if err != nil {
			return nil, err
		}
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		cache[strings.TrimSuffix(file.Name(), ".json")] = &cacheInfo{
			Timestamp: ft.UnixNano(),
			Data:      data,
		}
	}
	return cache, nil
}

func filePath(namespace string) string {
	return filepath.Join(dir, namespace)
}

func fileName(namespace, name string) string {
	return filepath.Join(dir, namespace, name+".json")
}

func fileTime(filename string) (time.Time, error) {
	stat, err := os.Stat(filename)
	if err != nil {
		return time.Time{}, err
	}

	return stat.ModTime(), nil
}

func getData(namespace, name string) ([]byte, error) {
	mutex.Lock()
	defer mutex.Unlock()

	filename := fileName(namespace, name)
	tm, err := fileTime(filename)
	if err != nil {
		return nil, err
	}

	var data []byte
	if cache[namespace] != nil {
		if info := cache[namespace][name]; info != nil {
			data = info.Data
			if tm.UnixNano() == info.Timestamp {
				return data, nil
			}
		}
	}

	file, err := ioutil.ReadFile(filename)
	if err != nil {
		if data != nil {
			return data, nil
		}
		return nil, err
	}

	if cache[namespace] == nil {
		cache[namespace] = make(map[string]*cacheInfo)
	}
	cache[namespace][name] = &cacheInfo{
		Data:      file,
		Timestamp: tm.UnixNano(),
	}

	return file, nil
}

func touch(filename string) (err error) {
	file, err := os.Create(filename)
	if err != nil {
		return
	}
	defer file.Close()

	return
}

func notify(namespace, name string, fn func()) (err error) {
	var last, curr time.Time
	for {
		filename := fileName(namespace, name)
		if curr, _ = fileTime(filename); curr != last {
			fn()
			last = curr
		}
		time.Sleep(time.Second)
	}
}
