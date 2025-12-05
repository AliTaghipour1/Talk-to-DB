package tracer

import (
	"sync"
	"time"
)

var once sync.Once
var tracer *Tracer

type Tracer struct {
	eventsChannel    chan *Event
	events           map[string]*keyData
	registeredEvents sync.Map
}

func newTracer() *Tracer {
	channel := make(chan *Event, 100000)
	result := &Tracer{
		eventsChannel: channel,
		events:        make(map[string]*keyData),
	}

	go result.run()
	return result
}

func GetTracer() *Tracer {
	once.Do(func() {
		tracer = newTracer()
	})
	return tracer
}

type Event struct {
	Key   string
	Value time.Duration
}

func (t *Tracer) registerEvent(key string) {
	t.registeredEvents.Store(key, "z")
	ticker := time.Tick(updateInterval)
	go func() {
		for range ticker {
			data, ok := t.events[key]
			if !ok {
				continue
			}
			data.resetAndLog()
		}
	}()
}

func (t *Tracer) isRegistered(key string) bool {
	_, ok := t.registeredEvents.Load(key)
	return ok
}

func (t *Tracer) SendEvent(event *Event) {
	if event == nil {
		return
	}
	if !t.isRegistered(event.Key) {
		t.registerEvent(event.Key)
	}

	t.eventsChannel <- event
}

func (t *Tracer) run() {
	for event := range t.eventsChannel {
		data, ok := t.events[event.Key]
		if !ok {
			data = &keyData{
				name:     event.Key,
				sumValue: event.Value.Milliseconds(),
				count:    1,
				start:    time.Now(),
			}
			t.events[event.Key] = data
		} else {
			data.addData(event.Value.Milliseconds())
		}
	}
}
