package agent

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/filipgorny/agent/core"
	"github.com/filipgorny/agent/message"
	"github.com/filipgorny/agent/plugins/shell"
	"github.com/filipgorny/agent/stream"
)

// drainMessages collects stream messages until stop is closed.
func drainMessages(a *Agent, stop <-chan struct{}) *[]stream.Record {
	var (
		mu  sync.Mutex
		got []stream.Record
	)

	out := &got

	go func() {
		for {
			select {

			case <-stop:
				return

			case m := <-a.Stream():
				mu.Lock()
				got = append(got, m)
				mu.Unlock()
			}
		}
	}()

	return out
}

func TestStreamEmitsDuringReason(t *testing.T) {
	strat := &scriptedLlm{
		reply: func(n int, prompt string) string {
			if n == 0 {
				return `{"action":"skill","params":{"name":"shell_run","command":"echo hi"}}`
			}

			return "done"
		},
	}

	a := newAgentWithLlm(t, strat, []core.Plugin{shell.ShellPlugin{}}, []string{"shell_run"}, "go")

	stop := make(chan struct{})
	got := drainMessages(a, stop)

	if _, err := a.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
	close(stop)

	types := map[string]bool{}

	for _, m := range *got {
		types[m.Type+"/"+m.Subtype] = true
	}

	for _, want := range []string{
		stream.TypeLog + "/" + stream.LogToolCall,
		stream.TypeLog + "/" + stream.LogToolResult,
		stream.TypeAnswerUser + "/",
	} {
		if !types[want] {
			t.Errorf("missing stream message %q; got %v", want, types)
		}
	}
}

func TestAskUserBlocksAndAnswers(t *testing.T) {
	a := newAgentWithLlm(t, &scriptedLlm{reply: func(int, string) string { return "" }}, nil, nil, "")
	a.SetInteractive(true)

	result := make(chan string, 1)

	go func() {
		ans, _ := a.askUser(context.Background(), stream.AskRequest{Question: "Which?", Choices: []string{"a", "b"}})
		result <- ans
	}()

	// Expect an ASK_USER message (CHOICE subtype).
	select {

	case m := <-a.Stream():

		if m.Type != stream.TypeAskUser || m.Subtype != stream.SubtypeChoice {
			t.Fatalf("expected ASK_USER/CHOICE, got %s/%s", m.Type, m.Subtype)
		}

	case <-time.After(time.Second):
		t.Fatal("no ASK_USER message")
	}

	a.Answer("b")

	select {

	case ans := <-result:

		if ans != "b" {
			t.Errorf("answer = %q, want b", ans)
		}

	case <-time.After(time.Second):
		t.Fatal("askUser did not return after Answer")
	}
}

func TestAskUserNonInteractive(t *testing.T) {
	a := newAgentWithLlm(t, &scriptedLlm{reply: func(int, string) string { return "" }}, nil, nil, "")

	if _, err := a.askUser(context.Background(), stream.AskRequest{Question: "Q"}); err == nil {
		t.Fatal("expected error in non-interactive mode")
	}
}

func TestChangeRoot(t *testing.T) {
	orig, _ := os.Getwd()

	defer os.Chdir(orig)

	dir := t.TempDir()
	a := newAgentWithLlm(t, &scriptedLlm{reply: func(int, string) string { return "" }}, nil, nil, "")

	stop := make(chan struct{})
	got := drainMessages(a, stop)

	if _, err := a.Execute(context.Background(), execContext{}, message.ActionCall{
		Action: ActionChangeRoot,
		Params: map[string]any{"path": dir},
	}); err != nil {
		t.Fatalf("change_root: %v", err)
	}

	time.Sleep(50 * time.Millisecond)
	close(stop)

	if a.Root() != dir {
		t.Errorf("root = %q, want %q", a.Root(), dir)
	}

	found := false

	for _, m := range *got {
		if m.Type == stream.TypeChangeRoot {
			found = true
		}
	}

	if !found {
		t.Error("no CHANGE_ROOT_FOLDER message emitted")
	}
}
