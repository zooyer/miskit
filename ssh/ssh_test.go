package ssh

import (
	"fmt"
	"testing"
)

func TestParse(t *testing.T) {
	var err error
	var user, addr, path string
	var tests = []struct {
		Input string
		User  string
		Addr  string
		Path  string
	}{
		{
			Input: "root@127.0.0.1:2022:/tmp/test",
			User:  "root",
			Addr:  "127.0.0.1:2022",
			Path:  "/tmp/test",
		},
		{
			Input: "127.0.0.1:/tmp/test",
			User:  "Administrator",
			Addr:  "127.0.0.1:22",
			Path:  "/tmp/test",
		},
		{
			Input: "root@127.0.0.1",
			User:  "root",
			Addr:  "127.0.0.1:22",
			Path:  "/",
		},
		{
			Input: "127.0.0.1:/",
			User:  "Administrator",
			Addr:  "127.0.0.1:22",
			Path:  "/",
		},
	}

	for _, test := range tests {
		if user, addr, path, err = parse(test.Input); err != nil {
			t.Fatal(err)
		}
		if user != test.User {
			t.Fatal(user, test.User)
		}
		if addr != test.Addr {
			t.Fatal(addr, test.Addr)
		}
		if path != test.Path {
			t.Fatal(path, test.Path)
		}
	}
}

func progress(current, total int64) {
	fmt.Println(current, total)
}

func TestScpFile(t *testing.T) {
	var err error
	if err = Scp("./攀D者-HC(2).mp4", "zzy@127.0.0.1:/home/zzy/video.mp4", "386143717", progress); err != nil {
		t.Fatal(err)
	}
}

func TestCmd(t *testing.T) {
	output, err := Command("zzy@127.0.0.1", "386143717", "ls")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(output)
}
