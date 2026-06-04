package domain

// Message represents a single dialogue message, aligned with proto Message.
type Message struct {
	Role        string     `json:"role"`
	Content     string     `json:"content"`
	ToolCallID  string     `json:"tool_call_id,omitempty"`
	IsTemporary bool       `json:"is_temporary,omitempty"`
	ToolCalls   []ToolCall `json:"tool_calls,omitempty"`
}

// ToolCall represents a function call requested by the LLM.
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
	Index    int          `json:"index,omitempty"`
}

// FunctionCall holds the function name and serialized arguments.
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ChatRequest is the internal representation of a StreamChat request.
type ChatRequest struct {
	SessionID         string    `json:"session_id"`
	RequestID         string    `json:"request_id"`
	DeviceID          string    `json:"device_id"`
	ClientID          string    `json:"client_id"`
	Query             string    `json:"query"`
	IntentType        string    `json:"intent_type"`
	MaxOutputSize     int       `json:"max_output_size"`
	Messages          []Message `json:"messages"`
	VoiceprintEnabled bool      `json:"voiceprint_enabled"`
}

// ChatChunk is a single piece of streaming response.
type ChatChunk struct {
	Event        string     `json:"event"`
	ContentChunk string     `json:"content_chunk,omitempty"`
	ToolCalls    []ToolCall `json:"tool_calls,omitempty"`
	FinishReason string     `json:"finish_reason,omitempty"`
	ErrorCode    string     `json:"error_code,omitempty"`
	ErrorMessage string     `json:"error_message,omitempty"`
}
