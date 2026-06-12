package interaction

import (
	"context"
	"testing"

	"github.com/filipgorny/agent/stream"
)

func TestAskUserSkill(t *testing.T) {
	var gotReq stream.AskRequest

	ask := func(ctx context.Context, req stream.AskRequest) (string, error) {
		gotReq = req

		return "the-answer", nil
	}

	out, err := AskUser{ask: ask}.Run(context.Background(), map[string]any{"question": "Q?"})

	if err != nil {
		t.Fatalf("run: %v", err)
	}

	if out != "the-answer" || gotReq.Question != "Q?" {
		t.Errorf("out=%q req=%+v", out, gotReq)
	}
}

func TestAskChoiceSkill(t *testing.T) {
	var gotReq stream.AskRequest

	ask := func(ctx context.Context, req stream.AskRequest) (string, error) {
		gotReq = req

		return req.Choices[1], nil
	}

	out, err := AskChoice{ask: ask}.Run(context.Background(), map[string]any{
		"question": "Pick",
		"choices":  []any{"x", "y", "z"},
	})

	if err != nil {
		t.Fatalf("run: %v", err)
	}

	if out != "y" || len(gotReq.Choices) != 3 {
		t.Errorf("out=%q req=%+v", out, gotReq)
	}
}

func TestAskRequiresInteractive(t *testing.T) {
	if _, err := (AskUser{}).Run(context.Background(), map[string]any{"question": "Q"}); err == nil {
		t.Error("ask_user without ask callback should error")
	}

	if _, err := (AskChoice{}).Run(context.Background(), map[string]any{"question": "Q", "choices": []any{"a"}}); err == nil {
		t.Error("ask_choice without ask callback should error")
	}
}
