package ui

import (
	"testing"

	"github.com/Curt-Park/claudeview/internal/model"
)

func TestBuildMergedChatItems_SingleSession(t *testing.T) {
	turns := []model.Turn{
		{Role: "user", Text: "hello"},
		{Role: "assistant", Text: "hi"},
	}
	items := BuildMergedChatItems([][]model.Turn{turns}, nil, nil, nil)
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].IsDivider || items[1].IsDivider {
		t.Error("single session should have no dividers")
	}
}

func TestBuildMergedChatItems_MultiSession(t *testing.T) {
	s1 := []model.Turn{
		{Role: "user", Text: "hello"},
		{Role: "assistant", Text: "hi"},
	}
	s2 := []model.Turn{
		{Role: "user", Text: "continue"},
		{Role: "assistant", Text: "sure"},
	}
	ids := []string{"aabbccdd", "eeffgghh"}
	items := BuildMergedChatItems([][]model.Turn{s1, s2}, nil, nil, ids)

	// s1(2) + divider(1) + s2(2) = 5
	if len(items) != 5 {
		t.Fatalf("expected 5 items, got %d", len(items))
	}
	if items[2].IsDivider != true {
		t.Error("expected divider at index 2")
	}
	expected := "── session 2/2 (eeffgghh) ──"
	if items[2].DividerLabel != expected {
		t.Errorf("expected divider label %q, got %q", expected, items[2].DividerLabel)
	}
}

func TestBuildMergedChatItems_ThreeSessions(t *testing.T) {
	s1 := []model.Turn{{Role: "user", Text: "a"}}
	s2 := []model.Turn{{Role: "user", Text: "b"}}
	s3 := []model.Turn{{Role: "user", Text: "c"}}
	ids := []string{"aaaa1111", "bbbb2222", "cccc3333"}
	items := BuildMergedChatItems([][]model.Turn{s1, s2, s3}, nil, nil, ids)

	// s1(1) + divider + s2(1) + divider + s3(1) = 5
	if len(items) != 5 {
		t.Fatalf("expected 5 items, got %d", len(items))
	}
	if !items[1].IsDivider {
		t.Error("expected divider at index 1")
	}
	if items[1].DividerLabel != "── session 2/3 (bbbb2222) ──" {
		t.Errorf("expected label with session ID, got %q", items[1].DividerLabel)
	}
	if !items[3].IsDivider {
		t.Error("expected divider at index 3")
	}
	if items[3].DividerLabel != "── session 3/3 (cccc3333) ──" {
		t.Errorf("expected label with session ID, got %q", items[3].DividerLabel)
	}
}

func TestBuildMergedChatItems_Empty(t *testing.T) {
	items := BuildMergedChatItems(nil, nil, nil, nil)
	if len(items) != 0 {
		t.Fatalf("expected 0 items for nil input, got %d", len(items))
	}
}

func TestBuildMergedChatItems_NilSessionIDs(t *testing.T) {
	s1 := []model.Turn{{Role: "user", Text: "a"}}
	s2 := []model.Turn{{Role: "user", Text: "b"}}
	items := BuildMergedChatItems([][]model.Turn{s1, s2}, nil, nil, nil)

	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}
	// Without session IDs, divider should not have parenthetical
	if items[1].DividerLabel != "── session 2/2 ──" {
		t.Errorf("expected label without session ID, got %q", items[1].DividerLabel)
	}
}

func TestDividerChatItem_Labels(t *testing.T) {
	d := ChatItem{IsDivider: true, DividerLabel: "── session 2/3 ──"}
	if d.WhoLabel() != "" {
		t.Errorf("expected empty WhoLabel for divider, got %q", d.WhoLabel())
	}
	if d.MessagePreview(120) != "── session 2/3 ──" {
		t.Errorf("expected DividerLabel in MessagePreview, got %q", d.MessagePreview(120))
	}
	if d.ActionLabel() != "" {
		t.Errorf("expected empty ActionLabel for divider, got %q", d.ActionLabel())
	}
	if d.ModelTokenLabel() != "" {
		t.Errorf("expected empty ModelTokenLabel for divider, got %q", d.ModelTokenLabel())
	}
	if d.TimeLabel(nil) != "" {
		t.Errorf("expected empty TimeLabel for divider, got %q", d.TimeLabel(nil))
	}
}

func TestBuildChatItems_SubagentIdxSet(t *testing.T) {
	mainTurns := []model.Turn{
		{Role: "user", Text: "hello"},
		{Role: "assistant", Text: "delegating", ToolCalls: []*model.ToolCall{{Name: "Task"}}},
		{Role: "assistant", Text: "delegating again", ToolCalls: []*model.ToolCall{{Name: "Task"}}},
	}
	sub0 := []model.Turn{
		{Role: "assistant", Text: "sub0-turn1"},
		{Role: "assistant", Text: "sub0-turn2"},
	}
	sub1 := []model.Turn{
		{Role: "assistant", Text: "sub1-turn1"},
	}
	subTypes := []model.AgentType{model.AgentTypeExplore, model.AgentTypePlan}
	items := BuildChatItems(mainTurns, [][]model.Turn{sub0, sub1}, subTypes)

	// Verify non-subagent items have SubagentIdx = -1
	for _, item := range items {
		if !item.IsSubagent && item.SubagentIdx != -1 {
			t.Errorf("non-subagent item %q should have SubagentIdx=-1, got %d", item.Turn.Text, item.SubagentIdx)
		}
	}

	// Collect subagent items and verify their SubagentIdx values
	var sub0Items, sub1Items []ChatItem
	for _, item := range items {
		if item.IsSubagent && item.SubagentIdx == 0 {
			sub0Items = append(sub0Items, item)
		}
		if item.IsSubagent && item.SubagentIdx == 1 {
			sub1Items = append(sub1Items, item)
		}
	}
	if len(sub0Items) != 2 {
		t.Errorf("expected 2 items with SubagentIdx=0, got %d", len(sub0Items))
	}
	if len(sub1Items) != 1 {
		t.Errorf("expected 1 item with SubagentIdx=1, got %d", len(sub1Items))
	}
	// Verify agent types are propagated
	for _, item := range sub0Items {
		if item.AgentType != model.AgentTypeExplore {
			t.Errorf("expected AgentType=Explore for sub0, got %v", item.AgentType)
		}
	}
	for _, item := range sub1Items {
		if item.AgentType != model.AgentTypePlan {
			t.Errorf("expected AgentType=Plan for sub1, got %v", item.AgentType)
		}
	}
}

func TestBuildChatItems_AgentToolName(t *testing.T) {
	// Real Claude Code transcripts use "Agent" (not "Task") as the tool name.
	mainTurns := []model.Turn{
		{Role: "user", Text: "hello"},
		{Role: "assistant", Text: "delegating", ToolCalls: []*model.ToolCall{{Name: "Agent"}}},
	}
	sub := []model.Turn{
		{Role: "assistant", Text: "agent-turn1"},
		{Role: "assistant", Text: "agent-turn2"},
	}
	items := BuildChatItems(mainTurns, [][]model.Turn{sub}, []model.AgentType{model.AgentTypeExplore})

	var subItems []ChatItem
	for _, item := range items {
		if item.IsSubagent {
			subItems = append(subItems, item)
		}
	}
	if len(subItems) != 2 {
		t.Fatalf("expected 2 subagent items for Agent tool call, got %d", len(subItems))
	}
	if subItems[0].SubagentIdx != 0 || subItems[1].SubagentIdx != 0 {
		t.Errorf("expected SubagentIdx=0 for both, got %d and %d", subItems[0].SubagentIdx, subItems[1].SubagentIdx)
	}
}
