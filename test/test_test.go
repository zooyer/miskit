/**
 * @Author: zzy
 * @Email: zhangzhongyuan@didiglobal.com
 * @Description:
 * @File: test_test.go
 * @Package: test
 * @Version: 1.0.0
 * @Date: 2021/12/29 4:47 下午
 */

package test

import (
	"fmt"
	"github.com/brahma-adshonor/gohook"
	"reflect"
	"strconv"
	"syscall"
	"testing"
	"unsafe"
)

func TestTTT(t *testing.T) {
	var f = 1.00
	addr := uintptr(unsafe.Pointer(&f))

	var i int64

	i = *(*int64)(unsafe.Pointer(addr))

	str := fmt.Sprintf("%064b", i)

	fmt.Println(str[:12])
	fmt.Println(str[12:])

	i3 := fmt.Sprintf("%064b", int64(3))
	fmt.Println(i3)
}

//go:linkname goexit1 runtime.goexit1
func goexit1()

//go:linkname runtime_beforeExit runtime.runtime_beforeExit
func runtime_beforeExit()

//go:linkname syscallExit syscall.Exit
func syscallExit()

//go:linkname exit runtime.exit
func exit(code int32)

//go:linkname main runtime.main
func main()

// ago:linkname os_beforeExit runtime.os_beforeExit
//go:linkname os_beforeExit os.runtime_beforeExit
func os_beforeExit()

func abc() {
	fmt.Println("aaa exit...")
}

//go:noinline
func aaa() {
	fmt.Println("aaa...")
}

//go:noinline
func ccc() {
	fmt.Println("ccc...")
}

//go:noinline
func bbb() {
	fmt.Println("bbb...")
}

func TestCeil2(t *testing.T) {

	var a = 3141592653.58979323846264338327950288419716939937510582097494459
	aa := strconv.FormatFloat(a, 'f', 60, 64)
	fmt.Println(aa)

	var f1 interface{} = syscallExit
	var f2 interface{} = syscall.Exit

	fmt.Println("f1:", f1)
	fmt.Println("f2:", f2)

	if reflect.DeepEqual(f1, f2) {
		fmt.Println("f1 == f2")
	} else {
		fmt.Println("f1 != f2")
	}

	if err := gohook.Hook(syscallExit, abc, syscallExit); err != nil {
		panic(err)
	}

	return
}

//go:noinline
func beforeExit() {
	fmt.Println("666 before exit...")
}

func hook(a, b, c interface{}) {
	if err := gohook.Hook(a, b, c); err != nil {
		panic(err)
	}
}

func TestHook(t *testing.T) {
	hook(aaa, bbb, ccc)
	aaa()
	bbb()
	ccc()
}

func TestHookExit(t *testing.T) {
	hook(os_beforeExit, beforeExit, nil)
}
