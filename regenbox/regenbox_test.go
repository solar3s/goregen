package regenbox

import (
	"testing"
	"time"
)

func testAutoConnect(t *testing.T) *RegenBox {
	rb, err := NewRegenBox(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	return rb
}

func TestRegenBox_LedToggle(t *testing.T) {
	rbx := testAutoConnect(t)
	t.Log("blinking led...")

	b0, err := rbx.LedToggle()
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 5; i++ {
		b, err := rbx.LedToggle()
		if err != nil {
			t.Fatal(err)
		}

		if b == b0 {
			t.Error("wrong return value for LedToggle()")
		}
		b0 = b
		time.Sleep(time.Millisecond * 500)
	}
}
