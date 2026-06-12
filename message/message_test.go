package message

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/filipgorny/agent/core"
)

var (
	_ InputMessage  = UserInput{}
	_ InputMessage  = EventMessage{}
	_ InputMessage  = ActionResult{}
	_ OutputMessage = ActionCall{}
)

func TestUserInputJSON(t *testing.T) {
	b, _ := json.Marshal(NewUserInput("hi"))
	s := string(b)

	if !strings.Contains(s, `"msg_type":"user_input"`) || !strings.Contains(s, `"text":"hi"`) {
		t.Errorf("bad user_input json: %s", s)
	}
}

func TestEventMessageJSON(t *testing.T) {
	ev := core.Event{Type: "file.changed", Source: "file_watch", ActionUID: "a1", ThreadID: "t1", Data: map[string]any{"path": "/x"}}
	b, _ := json.Marshal(NewEventMessage(ev))
	s := string(b)

	for _, want := range []string{`"msg_type":"event"`, `"event_type":"file.changed"`, `"action_uid":"a1"`, `"thread_id":"t1"`} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %s in %s", want, s)
		}
	}
}

func TestActionResultJSON(t *testing.T) {
	b, _ := json.Marshal(NewActionResult("skill", "ok"))
	s := string(b)

	if !strings.Contains(s, `"msg_type":"action_result"`) || !strings.Contains(s, `"result":"ok"`) {
		t.Errorf("bad action_result json: %s", s)
	}
}
