package server

import (
	"bytes"
	"strings"
	"testing"
)

func TestTransformSSELine_PassThroughChunk(t *testing.T) {
	// A minimal chunk similar to captured responses
	in := []byte(`{"id":"chatcmpl-xyz","object":"chat.completion.chunk","created":1754642367,"model":"gpt-4.1-2025-04-14","choices":[{"index":0,"delta":{"content":"Hello"},"logprobs":null,"finish_reason":null}]}`)
	out, done, err := TransformSSELine(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if done {
		t.Fatalf("done should be false for normal chunk")
	}
	if !bytes.Contains(out, []byte(`"chat.completion.chunk"`)) {
		t.Fatalf("expected object to remain chat.completion.chunk: %s", string(out))
	}
	if !bytes.Contains(out, []byte(`"Hello"`)) {
		t.Fatalf("expected content to be preserved: %s", string(out))
	}
}

func TestTransformSSELine_Done(t *testing.T) {
	out, done, err := TransformSSELine([]byte("[DONE]"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !done {
		t.Fatalf("expected done=true for [DONE]")
	}
	if len(out) != 0 {
		t.Fatalf("expected no output for [DONE], got: %s", string(out))
	}
}

func TestRewriteSSEStream_EndToEnd(t *testing.T) {
	src := strings.Join([]string{
		"data: {\"id\":\"chatcmpl-1\",\"object\":\"chat.completion.chunk\",\"created\":1,\"model\":\"gpt-4.1\",\"choices\":[{\"index\":0,\"delta\":{\"content\":\"Hi\"},\"logprobs\":null,\"finish_reason\":null}]}\n",
		"",
		"data: [DONE]",
		"",
	}, "\n")

	var dst bytes.Buffer
	if err := RewriteSSEStream(strings.NewReader(src), &dst, "gpt-5"); err != nil {
		t.Fatalf("rewrite failed: %v", err)
	}
	out := dst.String()
	// Expect two SSE events: one JSON, then DONE, each followed by blank line
	if !strings.Contains(out, "data: [DONE]\n\n") {
		t.Fatalf("missing DONE marker in output: %q", out)
	}
	if !strings.Contains(out, "\"content\":\"Hi\"") {
		t.Fatalf("missing transformed content in output: %q", out)
	}
}

func TestRewriteSSEStream_CodexToOpenAI(t *testing.T) {
	// Simulate a Codex backend SSE stream with created, two deltas, completed, DONE
	src := strings.Join([]string{
		"data: {\"type\":\"response.created\",\"sequence_number\":1,\"response\":{\"id\":\"resp_abc\"}}",
		"",
		"data: {\"type\":\"response.output_text.delta\",\"sequence_number\":2,\"delta\":\"Hello\"}",
		"",
		"data: {\"type\":\"response.output_text.delta\",\"sequence_number\":3,\"delta\":\" world\"}",
		"",
		"data: {\"type\":\"response.completed\",\"sequence_number\":4}",
		"",
		"data: [DONE]",
		"",
	}, "\n")

	var dst bytes.Buffer
	if err := RewriteSSEStream(strings.NewReader(src), &dst, "gpt-5"); err != nil {
		t.Fatalf("rewrite failed: %v", err)
	}
	out := dst.String()

	// Should include role chunk first
	if !strings.Contains(out, "\"delta\":{\"role\":\"assistant\"}") {
		t.Fatalf("missing assistant role delta: %q", out)
	}
	// Should include our content deltas
	if !strings.Contains(out, "\"content\":\"Hello\"") {
		t.Fatalf("missing first content delta: %q", out)
	}
	if !strings.Contains(out, "\"content\":\" world\"") {
		t.Fatalf("missing second content delta: %q", out)
	}
	// Should include a final stop chunk from response.completed
	if !strings.Contains(out, "\"finish_reason\":\"stop\"") {
		t.Fatalf("missing stop finish_reason: %q", out)
	}
	// And a DONE terminator
	if !strings.Contains(out, "data: [DONE]\n\n") {
		t.Fatalf("missing DONE marker: %q", out)
	}
}

func TestRewriteSSEStream_CodexToolCalls_ToOpenAI(t *testing.T) {
	src := strings.Join([]string{
		// created
		"data: {\"type\":\"response.created\",\"sequence_number\":1,\"response\":{\"id\":\"resp_tool\"}}",
		"",
		// tool call added
		"data: {\"type\":\"response.output_item.added\",\"sequence_number\":2,\"output_index\":0,\"item\":{\"id\":\"fc_123\",\"type\":\"function_call\",\"status\":\"in_progress\",\"arguments\":\"\",\"call_id\":\"call_abc\",\"name\":\"get_weather\"}}",
		"",
		// arguments streamed in parts
		`data: {"type":"response.function_call_arguments.delta","sequence_number":3,"item_id":"fc_123","output_index":0,"delta":"{\"location\":\"Pa"}`,
		"",
		`data: {"type":"response.function_call_arguments.delta","sequence_number":4,"item_id":"fc_123","output_index":0,"delta":"ris, France\"}"}`,
		"",
		// done and response completed
		"data: {\"type\":\"response.function_call_arguments.done\",\"sequence_number\":5,\"item_id\":\"fc_123\",\"output_index\":0}",
		"",
		"data: {\"type\":\"response.completed\",\"sequence_number\":6}",
		"",
		"data: [DONE]",
		"",
	}, "\n")

	var dst bytes.Buffer
	if err := RewriteSSEStream(strings.NewReader(src), &dst, "gpt-5"); err != nil {
		t.Fatalf("rewrite failed: %v", err)
	}
	out := dst.String()
	// Expect role chunk
	if !strings.Contains(out, "\"delta\":{\"role\":\"assistant\"}") {
		t.Fatalf("missing assistant role: %q", out)
	}
	// Expect initial tool_calls with function name and id
	if !strings.Contains(out, "\"tool_calls\"") || !strings.Contains(out, "\"name\":\"get_weather\"") {
		t.Fatalf("missing tool_calls start/name: %q", out)
	}
	if !strings.Contains(out, "\"id\":\"call_abc\"") {
		t.Fatalf("missing tool call id: %q", out)
	}
	// Expect arguments deltas streamed
	if !strings.Contains(out, "\"arguments\":\"{\\\"location\\\":\\\"Pa") {
		t.Fatalf("missing first args delta: %q", out)
	}
	// Be lenient about escaping differences; accept common encodings
	secondOk := strings.Contains(out, "\"arguments\":\"ris, France\\\"}\"") ||
		strings.Contains(out, "\"arguments\":\"ris, France\\\\\\\"}\\\"\"") ||
		strings.Contains(out, "\"arguments\":\"ris, France\\\"}\"")
	if !secondOk {
		t.Fatalf("missing second args delta: %q", out)
	}
	// Finish reason should be tool_calls
	if !strings.Contains(out, "\"finish_reason\":\"tool_calls\"") {
		t.Fatalf("missing tool_calls finish reason: %q", out)
	}
	if !strings.Contains(out, "data: [DONE]\n\n") {
		t.Fatalf("missing DONE marker: %q", out)
	}
}
