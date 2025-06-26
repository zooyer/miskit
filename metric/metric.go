package metric

import (
	"fmt"
	"time"
)

var debug bool

func Rpc(name, caller, callee string, code int, latency time.Duration, tag map[string]interface{}) {
	if debug {
		fmt.Println("[METRIC - RPC]", time.Now(), name, caller, callee, code, latency, tag)
	}
}

func Count(name string, count int, tag map[string]interface{}) {
	if debug {
		fmt.Println("[METRIC - COUNT]", time.Now(), name, count, tag)
	}
}

func SetDebug(dbg bool) {
	debug = dbg
}
