/**
 * @Author: zzy
 * @Email: zhangzhongyuan@didiglobal.com
 * @Description:
 * @File: math_test.go
 * @Package: maths
 * @Version: 1.0.0
 * @Date: 2021/12/15 8:41 下午
 */

package math

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCeil(t *testing.T) {
	assert.Equal(t, Ceil(1.23456, 0), 2.0)
	assert.Equal(t, Ceil(1.23456, 1), 1.3)
	assert.Equal(t, Ceil(1.23456, 2), 1.24)
	assert.Equal(t, Ceil(1.23456, 3), 1.235)
	assert.Equal(t, Ceil(1.23456, 4), 1.2346)
	assert.Equal(t, Ceil(1.23456, 5), 1.23457)
	assert.Equal(t, Ceil(1.23456, 6), 1.23456)
}

func TestFloor(t *testing.T) {
	assert.Equal(t, Floor(1.23456, 0), 1.0)
	assert.Equal(t, Floor(1.23456, 1), 1.2)
	assert.Equal(t, Floor(1.23456, 2), 1.23)
	assert.Equal(t, Floor(1.23456, 3), 1.234)
	assert.Equal(t, Floor(1.23456, 4), 1.2345)
	assert.Equal(t, Floor(1.23456, 5), 1.23456)
	assert.Equal(t, Floor(1.23456, 6), 1.23456)
}

func TestRound(t *testing.T) {
	assert.Equal(t, Round(1.23456, 0), 1.0)
	assert.Equal(t, Round(1.23456, 1), 1.2)
	assert.Equal(t, Round(1.23456, 2), 1.23)
	assert.Equal(t, Round(1.23456, 3), 1.235)
	assert.Equal(t, Round(1.23456, 4), 1.2346)
	assert.Equal(t, Round(1.23456, 5), 1.23456)
	assert.Equal(t, Round(1.23456, 6), 1.23456)
}
