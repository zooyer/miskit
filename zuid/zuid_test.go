package zuid

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/bwmarrin/snowflake"
)

func TestNewNode(t *testing.T) {
	sf, err := snowflake.NewNode(1)
	if err != nil {
		panic(err)
	}

	yf := Snowflake(snowflake.Epoch, 1)

	sid := sf.Generate()
	yid := yf.GenID()
	if sid.Int64() != yid.Int64() {
		t.Fatal("generate fail", sid, ":", yid)
	}
	fmt.Println(sid, ":", yid)

	sid = sf.Generate()
	yid = yf.GenID()
	if sid.Int64() != yid.Int64() {
		t.Fatal("generate fail", sid, ":", yid)
	}
	fmt.Println(sid, ":", yid)

	shot := Shotflake(snowflake.Epoch, 1)
	now := time.Now()
	for i := 0; i < 16; i++ {
		fmt.Println(shot.GenID())
	}
	cost := time.Since(now)
	if seconds := cost.Seconds(); seconds < 2 || seconds > 3 {
		t.Fatal("generate fail", cost)
	}
	fmt.Println(time.Since(now))

	var id, id2 ID
	data, err := json.Marshal(id)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &id2); err != nil {
		t.Fatal(err)
	}
	if id != id2 {
		t.Fatal("json marshal error:", id, id2)
	}
}
