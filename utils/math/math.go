/**
 * @Author: zzy
 * @Email: zhangzhongyuan@didiglobal.com
 * @Description:
 * @File: math.go
 * @Package: maths
 * @Version: 1.0.0
 * @Date: 2021/12/15 7:46 下午
 */

package math

import (
	"math"
)

// Ceil 向上取整, 保留n位小数
func Ceil(f float64, n int) float64 {
	pow := math.Pow10(n)
	ceil := math.Ceil(f * pow)
	return math.Trunc(ceil) / pow
}

// Floor 向下取整, 保留n位小数
func Floor(f float64, n int) float64 {
	pow := math.Pow10(n)
	floor := math.Floor(f * pow)
	return math.Trunc(floor) / pow
}

// Round 四舍五入, 保留n位小数
func Round(f float64, n int) float64 {
	pow := math.Pow10(n)
	round := math.Round(f * pow)
	return math.Trunc(round) / pow
}
