package agent

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestMemoryActions(t *testing.T) {
	a := newAgentWithLlm(t, &scriptedLlm{reply: func(int, string) string { return "" }}, nil, "")
	ctx := context.Background()

	if _, err := a.Execute(ctx, ActionCall{Action: ActionRemember, Params: map[string]any{
		"content": "Paris is the capital of France",
	}}); err != nil {
		t.Fatalf("remember: %v", err)
	}

	out, err := a.Execute(ctx, ActionCall{Action: ActionRead, Params: map[string]any{
		"query": "capital France",
	}})

	if err != nil {
		t.Fatalf("read: %v", err)
	}

	if !strings.Contains(out, "Paris") {
		t.Errorf("read result missing remembered fact: %q", out)
	}
}

// signalLlm records each prompt it receives and always returns a shell action.
type signalLlm struct {
	got chan string
}

func (s *signalLlm) Prompt(ctx context.Context, prompt string) (string, error) {
	select {

	case s.got <- prompt:

	default:
	}

	return `{"action":"skill","params":{"name":"shell_run","command":"echo hi"}}`, nil
}

func TestListenReactsToEvent(t *testing.T) {
	sl := &signalLlm{got: make(chan string, 8)}
	a := newAgentWithLlm(t, sl, []string{"shell_run"}, "")

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	go func() {
		_ = a.Listen(ctx)
	}()

	deadline := time.After(3 * time.Second)
	tick := time.NewTicker(50 * time.Millisecond)

	defer tick.Stop()

	for {
		select {

		case p := <-sl.got:

			if strings.Contains(p, `"msg_type":"event"`) {
				return
			}

		case <-tick.C:
			a.Bus().Publish(Event{Type: "file.changed", Source: "file_watch", Data: map[string]any{"path": "/x"}})

		case <-deadline:
			t.Fatal("agent did not react to event")
		}
	}
}

func TestPreambleLanguageAndProtocol(t *testing.T) {
	a := newAgentWithLlm(t, &scriptedLlm{reply: func(int, string) string { return "" }}, []string{"shell_run"}, "")
	pre := a.protocolPreamble()

	for _, want := range []string{"msg_type", "user_input", "event", ActionRemember, ActionRead, ActionSkill, "shell_run", "English"} {
		if !strings.Contains(pre, want) {
			t.Errorf("preamble missing %q:\n%s", want, pre)
		}
	}

	// Configurable language.
	pl := NewAgent(a.llm, nil, NewEventBus(), NewInMemoryMemory(), "", "Polish")

	if !strings.Contains(pl.protocolPreamble(), "Polish") {
		t.Error("preamble should mention Polish")
	}
}
