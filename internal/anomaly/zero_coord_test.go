package anomaly

import (
	"testing"
	"time"
)

func TestDetector_ZeroCoordsNoFalseLoop(t *testing.T) {
	tripped := false
	d := NewDetector(func(tr Trip) { tripped = true })
	for i := 0; i < 10; i++ {
		d.Record("chat", 0, 0, true)
	}
	time.Sleep(50 * time.Millisecond)
	if tripped {
		t.Fatal("zero coordinates must not trigger loop detection")
	}
}
