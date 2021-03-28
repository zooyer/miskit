package conf

import (
	"os"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	const (
		filename = "./test/test.yaml"
	)
	var (
		err  error
		name string
		age  int
	)

	if err = os.WriteFile(filename, []byte("name: \"tom\"\nage: 18"), 0644); err != nil {
		t.Fatal(err)
	}

	conf := New(time.Millisecond*500)
	if err = conf.Init(filename); err != nil {
		t.Fatal(err)
	}
	if err = conf.Bind("name", &name); err != nil {
		t.Fatal(err)
	}
	if err = conf.Bind("age", &age); err != nil {
		t.Fatal(err)
	}
	t.Log(name, age)

	if err = os.WriteFile(filename, []byte("name: \"jack\"\nage: 15"), 0644); err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second)
	t.Log(name, age)
	return
}
