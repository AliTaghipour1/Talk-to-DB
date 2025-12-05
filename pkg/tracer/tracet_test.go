package tracer

import (
	"testing"
	"time"
)

func TestTracer(t *testing.T) {
	tr := GetTracer()

	for i := 0; i < 1000; i++ {
		go func() {
			keyName := "odd-event"
			if i%2 == 0 {
				keyName = "even-event"
			}
			tr.SendEvent(&Event{
				Key:   keyName,
				Value: time.Duration(i) * time.Millisecond,
			})
		}()
	}

	time.Sleep(12 * time.Second)
}
