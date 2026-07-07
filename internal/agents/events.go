package agents

import (
	"encoding/json"
	"strings"
)

type TerminalEventKind string

const (
	TerminalEventRaw       TerminalEventKind = "raw"
	TerminalEventMessage   TerminalEventKind = "message"
	TerminalEventToolCall  TerminalEventKind = "tool_call"
	TerminalEventUsage     TerminalEventKind = "usage"
	TerminalEventCompleted TerminalEventKind = "completed"
	TerminalEventError     TerminalEventKind = "error"
)

type TerminalEvent struct {
	ProviderID string            `json:"provider_id"`
	Kind       TerminalEventKind `json:"kind"`
	Text       string            `json:"text,omitempty"`
	Raw        string            `json:"raw,omitempty"`
	Metadata   map[string]any    `json:"metadata,omitempty"`
}

func ParseTerminalEvent(providerID, line string) TerminalEvent {
	event := TerminalEvent{
		ProviderID: normalizeProviderID(providerID),
		Kind:       TerminalEventRaw,
		Raw:        line,
		Text:       line,
	}
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || !strings.HasPrefix(trimmed, "{") {
		return event
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(trimmed), &payload); err != nil {
		return event
	}
	event.Metadata = payload
	event.Text = firstJSONText(payload, "message", "text", "content", "summary", "error")
	event.Kind = classifyTerminalEvent(payload)
	return event
}

func classifyTerminalEvent(payload map[string]any) TerminalEventKind {
	value := strings.ToLower(firstJSONText(payload, "type", "event", "kind", "status"))
	switch {
	case strings.Contains(value, "tool"):
		return TerminalEventToolCall
	case strings.Contains(value, "usage") || strings.Contains(value, "token") || strings.Contains(value, "cost"):
		return TerminalEventUsage
	case strings.Contains(value, "complete") || strings.Contains(value, "done") || strings.Contains(value, "success"):
		return TerminalEventCompleted
	case strings.Contains(value, "error") || strings.Contains(value, "fail"):
		return TerminalEventError
	case value != "":
		return TerminalEventMessage
	default:
		return TerminalEventMessage
	}
}

func firstJSONText(payload map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := payload[key]; ok {
			switch typed := value.(type) {
			case string:
				if strings.TrimSpace(typed) != "" {
					return typed
				}
			case map[string]any:
				if text := firstJSONText(typed, keys...); text != "" {
					return text
				}
			}
		}
	}
	return ""
}
