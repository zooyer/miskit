package zuid

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/zooyer/jsons"
	"io"
	"net/http"
	"net/url"
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

func TestTTT(t *testing.T) {
	var params = url.Values{
		"app_id":     {"622660"},
		"app_secret": {"d566d332f9c78479baba83a6af7f6c45"},
	}
	url, err := url.Parse("http://ice.zhangzhongyuan.online:8004/ice/v1/idgen")
	if err != nil {
		t.Fatal(err)
	}
	url.RawQuery = params.Encode()
	const ur = "http://ice.zhangzhongyuan.online:8004/ice/v1/idgen?app_id=622660&app_secret=d566d332f9c78479baba83a6af7f6c45"
	GenID := func() (id int64, err error) {
		resp, err := http.Get(url.String())
		if err != nil {
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return 0, errors.New(resp.Status)
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return
		}

		val, err := jsons.Unmarshal(data)
		if err != nil {
			return
		}

		if val.Int("errno") != 0 {
			return 0, errors.New(val.String("errmsg"))
		}

		return val.Int("data", "id"), nil
	}

	var now = time.Now()

	for i := 0; i < 16; i++ {
		id, err := GenID()
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(id)
	}

	t.Log(time.Since(now))
}
