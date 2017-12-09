package regenbox

import (
	"sync"
	"testing"
	"time"
)

var rb *RegenBox
var once sync.Once

func testAutoConnect(tb testing.TB) *RegenBox {
	if rb != nil {
		return rb
	}

	var err error
	rb, err = NewRegenBox(nil, nil)
	if err != nil {
		tb.Skip(err)
	}
	return rb
}

func TestRegenBox_LedToggle(t *testing.T) {
	rbx := testAutoConnect(t)
	defer rbx.Stop()

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

func BenchmarkReadVoltage(b *testing.B) {
	rbx := testAutoConnect(b)
	defer rbx.Stop()

	once.Do(func() {
		b.Logf("using firmware: %s", rbx.FirmwareVersion())
	})

	var total int
	for n := 0; n < b.N; n++ {
		i, err := rbx.ReadVoltage()
		time.Sleep(time.Millisecond * 2)
		if err != nil {
			b.Error(b.N, err)
		}
		total += i
	}
	b.Logf("b.N: %d", b.N)
	b.Logf("average voltage: %d", total/b.N)
}
