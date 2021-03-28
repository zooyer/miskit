package zuid

import (
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
	fmt.Println(sid)
	fmt.Println(yid)

	sid = sf.Generate()
	yid = yf.GenID()
	fmt.Println(sid)
	fmt.Println(yid)

	shot := Shotflake(snowflake.Epoch, 1)
	now := time.Now()
	for i := 0; i < 16; i++ {
		fmt.Println(shot.GenID())
	}
	fmt.Println(time.Since(now))
}
