package model

import "encoding/json"

// ToolCallInfo holds the minimal fields needed to extract subagent types,
// allowing a single implementation to work with both model and transcript turn types.
type ToolCallInfo struct {
	Name  string
	Input json.RawMessage
}

// ExtractAgentTypesFromCalls reads Agent/Task tool calls and returns the subagent_type
// value for each, in call order. This matches the positional order used by
// BuildChatItems to interleave subagent turns.
func ExtractAgentTypesFromCalls(calls []ToolCallInfo) []AgentType {
	var types []AgentType
	for _, c := range calls {
		if c.Name != "Agent" && c.Name != "Task" {
			continue
		}
		types = append(types, AgentTypeFromInput(c.Input))
	}
	return types
}

// ExtractSubagentTypes collects Agent/Task tool calls from Turn slices
// and delegates to ExtractAgentTypesFromCalls.
func ExtractSubagentTypes(turns []Turn) []AgentType {
	var calls []ToolCallInfo
	for _, t := range turns {
		if t.Role != "assistant" {
			continue
		}
		for _, tc := range t.ToolCalls {
			calls = append(calls, ToolCallInfo{Name: tc.Name, Input: tc.Input})
		}
	}
	return ExtractAgentTypesFromCalls(calls)
}
