package errors

import (
	"encoding/json"
	"testing"
)

func TestTTT(t *testing.T) {
	var a = map[string]interface{}{
		"name": "zs",
		"age":  24,
	}

	var b = make(map[interface{}]interface{})

	data, _ := json.Marshal(a)

	if err := json.Unmarshal(data, &b); err != nil {
		t.Fatal(err)
	}

	t.Logf("%#v", b)
}
