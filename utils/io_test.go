package utils

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestAppendFile(t *testing.T) {
	var err error
	const name = "test.txt"
	defer os.Remove(name)

	if err = Append(name, []byte("abc\n")); err != nil {
		t.Fatal(err)
	}
	if err = Append(name, []byte("def\n")); err != nil {
		t.Fatal(err)
	}
	if err = Append(name, []byte("ok\n")); err != nil {
		t.Fatal(err)
	}
	data, err := ioutil.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "abc\ndef\nok\n" {
		t.Fatal("append file fail")
	}
}
