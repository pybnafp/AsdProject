package common

import (
	"math/rand"
	"time"
)

// RandomInt returns a random integer between min and max, inclusive.
func RandomInt(min, max int) int {
	if min > max {
		min, max = max, min
	}
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min+1) + min
}
