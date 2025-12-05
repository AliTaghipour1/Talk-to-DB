package tracer

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type keyData struct {
	name     string
	sumValue int64
	count    int
	maxValue int64
	start    time.Time
	lock     sync.Mutex
}

const updateInterval = 20 * time.Second

func (k *keyData) addData(value int64) {
	k.lock.Lock()
	defer k.lock.Unlock()
	k.sumValue += value
	k.maxValue = max(k.maxValue, value)
	k.count++
}
func (k *keyData) resetAndLog() {
	k.lock.Lock()
	defer k.lock.Unlock()
	if k.count > 0 {
		log.Println(k.string())
	}
	k.sumValue = 0
	k.count = 0
	k.maxValue = 0
	k.start = time.Now()
}

const logMessageFormat = "[TRACER] - keyName: [%s] - averageValue: [%d ms] - maxValue: [%d ms] - throughput: [%.2f/s]"

func (k *keyData) string() string {
	averageValue := int(k.sumValue / max(int64(k.count), 1))
	throughput := float64(k.count) / time.Since(k.start).Seconds()
	return fmt.Sprintf(logMessageFormat, k.name, averageValue, k.maxValue, throughput)
}
