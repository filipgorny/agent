package core

import (
	"testing"
	"time"
)

func TestEventBusPublishSubscribe(t *testing.T) {
	bus := NewEventBus()
	ch, unsubscribe := bus.Subscribe()

	bus.Publish(Event{Type: "x", Source: "test"})

	select {

	case ev := <-ch:

		if ev.Type != "x" {
			t.Errorf("event type = %q", ev.Type)
		}

	case <-time.After(time.Second):
		t.Fatal("did not receive event")
	}

	unsubscribe()

	if _, ok := <-ch; ok {
		t.Error("channel should be closed after unsubscribe")
	}
}

func TestEventBusClose(t *testing.T) {
	bus := NewEventBus()
	ch, _ := bus.Subscribe()

	bus.Close()

	if _, ok := <-ch; ok {
		t.Error("channel should be closed after bus close")
	}

	bus.Publish(Event{Type: "y"})
}
