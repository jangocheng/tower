package events

import (
	"fmt"

	"github.com/admpub/events/meta"
)

const (
	//ModeAsync async
	ModeAsync = iota

	//ModeSync sync
	ModeSync

	//ModeWait async & sync.Wait
	ModeWait
)

type Emitter interface {
	On(string, ...Listener) Emitter //AddEventListener
	Off(string) Emitter             //RemoveEventListeners
	Fire(interface{}, int, ...meta.Map) error
	Events() map[string]Dispatcher
	HasEvent(string) bool
}

type Dispatcher interface {
	AddSubscribers(...Listener)
	Dispatch(Event) error
}

type DispatcherFactory func() Dispatcher

type Listener interface {
	Handle(Event) error
}

type Stream chan Event

func (stream Stream) Handle(event Event) error {
	stream <- event
	return nil
}

type Callback func(Event) error

func (callback Callback) Handle(event Event) error {
	return callback(event)
}

func New(name string) Event {
	return Event{
		Key:     name,
		Context: meta.Map{},
	}
}

type Event struct {
	Key     string
	Context meta.Map
	aborted bool
}

func (event *Event) String() string {
	return event.Key
}

func (event *Event) Abort() *Event {
	event.aborted = true
	return event
}

func (event *Event) Aborted() bool {
	return event.aborted
}

func ToMap(key string, value interface{}, args ...interface{}) meta.Map {
	context := meta.Map{key: value}
	for i, j := 0, len(args); i < j; i++ {
		if i%2 == 0 {
			key = fmt.Sprint(args[i])
			continue
		}
		context[key] = args[i]
	}
	return context
}
