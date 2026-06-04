package service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"xiaozhi-go-core/internal/pb/chatv1"
)

// ToolService provides a minimal ExecuteTool implementation.
// Current strategy: gateway keeps tool execution responsibility, so Go returns NOTFOUND.
type ToolService struct{}

func NewToolService() *ToolService { return &ToolService{} }

func (s *ToolService) ExecuteTool(req *chatv1.ExecuteToolRequest) *chatv1.ExecuteToolResponse {
	args := map[string]interface{}{}
	if req.ArgumentsJson != "" {
		_ = json.Unmarshal([]byte(req.ArgumentsJson), &args)
	}

	// Whitelist tools for minimal executable path in Go.
	switch req.ToolName {
	case "echo_text":
		text, _ := args["text"].(string)
		return &chatv1.ExecuteToolResponse{
			Action:   chatv1.ToolAction_TOOL_ACTION_RESPONSE,
			Response: text,
			Result:   text,
		}
	case "time_now":
		format, _ := args["format"].(string)
		if format == "" {
			format = time.RFC3339
		}
		now := time.Now().Format(format)
		return &chatv1.ExecuteToolResponse{
			Action:   chatv1.ToolAction_TOOL_ACTION_RESPONSE,
			Response: now,
			Result:   now,
		}
	case "math_add":
		a, errA := toFloat64(args["a"])
		b, errB := toFloat64(args["b"])
		if errA != nil || errB != nil {
			return &chatv1.ExecuteToolResponse{
				Action:       chatv1.ToolAction_TOOL_ACTION_ERROR,
				ErrorCode:    "TOOL_INVALID_ARGUMENT",
				ErrorMessage: "math_add requires numeric a and b",
			}
		}
		sum := a + b
		result := fmt.Sprintf("%v", sum)
		return &chatv1.ExecuteToolResponse{
			Action:   chatv1.ToolAction_TOOL_ACTION_RESPONSE,
			Response: result,
			Result:   result,
		}
	}

	return &chatv1.ExecuteToolResponse{
		Action:       chatv1.ToolAction_TOOL_ACTION_NOTFOUND,
		ErrorCode:    "TOOL_NOT_IMPLEMENTED_IN_GO",
		ErrorMessage: "tool execution stays in python gateway",
	}
}

func toFloat64(v interface{}) (float64, error) {
	switch n := v.(type) {
	case float64:
		return n, nil
	case float32:
		return float64(n), nil
	case int:
		return float64(n), nil
	case int64:
		return float64(n), nil
	case json.Number:
		return n.Float64()
	case string:
		return strconv.ParseFloat(n, 64)
	default:
		return 0, fmt.Errorf("not a number: %T", v)
	}
}
