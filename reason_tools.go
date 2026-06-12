package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/filipgorny/agent/message"
	"github.com/filipgorny/agent/runtime"
	llm "github.com/filipgorny/llm-provider"
)

// drive runs a goal to a conclusion, using native tool calling when the strategy
// supports it. If the model turns out not to support tools at runtime, it falls
// back to the prompt-based reasoning loop.
func (a *Agent) drive(ctx context.Context, threadID string, goal message.InputMessage) (string, error) {
	if tc, ok := a.llm.Strategy().(llm.ToolCaller); ok {
		out, err := a.reasonWithTools(ctx, tc, threadID, goal)

		if err != nil && errors.Is(err, llm.ErrToolsUnsupported) {
			return a.reason(ctx, threadID, goal)
		}

		return out, err
	}

	return a.reason(ctx, threadID, goal)
}

// reasonWithTools drives the goal via native tool calling: the model receives
// tool schemas and returns structured tool calls, which the agent executes and
// feeds back as tool messages, until the model replies with a plain-text answer.
func (a *Agent) reasonWithTools(ctx context.Context, tc llm.ToolCaller, threadID string, goal message.InputMessage) (string, error) {
	tools := a.buildToolSpecs()
	goalJSON, _ := json.Marshal(goal)

	messages := []llm.Message{
		{Role: "system", Content: a.toolSystemPrompt()},
		{Role: "user", Content: "Goal: " + string(goalJSON)},
	}

	for step := 0; step < a.maxSteps; step++ {
		resp, err := tc.CallTools(ctx, messages, tools)

		if err != nil {
			return "", fmt.Errorf("agent: tool reason: %w", err)
		}

		if len(resp.Calls) == 0 {
			return strings.TrimSpace(resp.Text), nil
		}

		messages = append(messages, llm.Message{Role: "assistant", Content: summarizeCalls(resp.Calls)})

		for _, call := range resp.Calls {
			ac := message.ActionCall{Action: call.Name, Params: call.Arguments}

			result, err := a.Execute(ctx, execContext{threadID: threadID, actionUID: runtime.NewUID()}, ac)

			if err != nil {
				result = "error: " + err.Error()
			}

			messages = append(messages, llm.Message{Role: "tool", Content: call.Name + " -> " + a.condense(result)})
		}

		messages = a.boundMessages(messages)
	}

	return "", fmt.Errorf("agent: tool reasoning did not conclude in %d steps", a.maxSteps)
}

// toolSystemPrompt is the system message for native tool-calling mode (format is
// handled by the tools API, so it only conveys behavior and language).
func (a *Agent) toolSystemPrompt() string {
	return fmt.Sprintf("You are an autonomous agent. Use the provided tools to investigate and act, "+
		"then give a final answer as plain text (no tool call). Do not repeat a tool call whose result "+
		"you already have. Always respond in %s.", a.language)
}

func summarizeCalls(calls []llm.ToolCall) string {
	parts := make([]string, 0, len(calls))

	for _, c := range calls {
		args, _ := json.Marshal(c.Arguments)
		parts = append(parts, c.Name+"("+string(args)+")")
	}

	return "calling: " + strings.Join(parts, ", ")
}

// boundMessages keeps the system message, the goal and a bounded window of the
// most recent turns, so context stays small for limited-context models.
func (a *Agent) boundMessages(msgs []llm.Message) []llm.Message {
	maxMsgs := 2 + maxStepsInContext*2

	if len(msgs) <= maxMsgs {
		return msgs
	}

	out := make([]llm.Message, 0, maxMsgs)
	out = append(out, msgs[0], msgs[1])
	out = append(out, msgs[len(msgs)-(maxMsgs-2):]...)

	return out
}
