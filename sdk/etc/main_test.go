package etc

import (
	"fmt"
	"testing"
	"time"

	"github.com/zooyer/jsons"
)

func init() {
	var err error
	if err = Init("./test"); err != nil {
		panic(err)
	}
}

func TestGetConfig(t *testing.T) {
	var err error
	var raw jsons.Raw
	if err = GetConfig("test", "test", &raw); err != nil {
		t.Fatal(err)
	}

	t.Log(string(raw))

	for i := 0; i < 60; i++ {
		var raw []byte
		if raw, err = GetData("test", "abc"); err != nil {
			t.Fatal(err)
		}
		fmt.Println(string(raw))
		time.Sleep(time.Second)
	}
}

func TestNotify(t *testing.T) {
	var err error
	fn := func() {
		data, err := GetData("test", "def")
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println(string(data))
	}
	if err = Notify("test", "def", fn); err != nil {
		t.Fatal(err)
	}
}
