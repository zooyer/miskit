package math

import (
	"math"
	"testing"
)

func TestRound(t *testing.T) {
	t.Log(math.Round(-12345.6))
	t.Log(math.Round(12345.6))
	t.Log(Round(-12345.6, 0))
	t.Log(Round(12345.6, 0))

	t.Log()

	t.Log(math.Floor(-12345.6))
	t.Log(math.Floor(12345.6))
	t.Log(Floor(-12345.6, 0))
	t.Log(Floor(12345.6, 0))

	t.Log()

	t.Log(math.Ceil(-12345.6))
	t.Log(math.Ceil(12345.6))
	t.Log(Ceil(-12345.6, 0))
	t.Log(Ceil(12345.6, 0))
}
