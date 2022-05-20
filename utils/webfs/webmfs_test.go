package webfs

import (
	"io/ioutil"
	"testing"
)

func TestWebMFS(t *testing.T) {
	var fs = WebMFS(map[string][]byte{
		"test": []byte("hello,world"),
	})

	file, err := fs.Open("test")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(data))

	data, err = ioutil.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(data))
}
