package agent

import (
	"encoding/json"
	"strings"
	"testing"
)

// compile-time checks: input/output message membership.
var (
	_ InputMessage  = UserInput{}
	_ InputMessage  = EventMessage{}
	_ OutputMessage = ActionCall{}
)

func TestUserInputJSON(t *testing.T) {
	b, err := json.Marshal(NewUserInput("hi"))

	if err != nil {
		t.Fatal(err)
	}

	s := string(b)

	if !strings.Contains(s, `"msg_type":"user_input"`) {
		t.Errorf("missing msg_type: %s", s)
	}

	if !strings.Contains(s, `"text":"hi"`) {
		t.Errorf("missing text: %s", s)
	}
}

func TestEventMessageJSON(t *testing.T) {
	b, err := json.Marshal(NewEventMessage("file.changed", "file_watch", map[string]any{"path": "/x"}))

	if err != nil {
		t.Fatal(err)
	}

	s := string(b)

	for _, want := range []string{`"msg_type":"event"`, `"event_type":"file.changed"`, `"source":"file_watch"`} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %s in %s", want, s)
		}
	}
}
