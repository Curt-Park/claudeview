# Tree Connectors for Sub-agent Rows — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add `├─` / `└─` tree connectors to the NAME column so each sub-agent row shows its parent-child relationship to the Claude turn that spawned it.

**Architecture:** Add a `TreeConnector` field to `ChatItem`; populate it in `interleaveSubagents` by pre-counting Agent/Task calls per batch and assigning `├─` to all but the last and `└─` to the last; `WhoLabel()` prepends the connector to the agent display name. Widen the NAME column from 10 to 14.

**Tech Stack:** Go, Bubble Tea, lipgloss; tests with `testing` stdlib.

---

### Task 1: Add `TreeConnector` field and update `WhoLabel()`

**Files:**
- Modify: `internal/ui/chat_item.go`
- Test: `internal/ui/chat_item_test.go`

**Step 1: Write the failing test**

Add to `internal/ui/chat_item_test.go`:

```go
func TestWhoLabel_TreeConnector(t *testing.T) {
	mid := ui.ChatItem{
		IsSubagent:    true,
		AgentType:     model.AgentTypeExplore,
		SubagentIdx:   0,
		TreeConnector: "├─",
	}
	if got := mid.WhoLabel(); got != "├─ Explorer" {
		t.Errorf("expected '├─ Explorer', got %q", got)
	}

	last := ui.ChatItem{
		IsSubagent:    true,
		AgentType:     model.AgentTypePlan,
		SubagentIdx:   1,
		TreeConnector: "└─",
	}
	if got := last.WhoLabel(); got != "└─ Planner" {
		t.Errorf("expected '└─ Planner', got %q", got)
	}

	// Sub-agent with no connector (empty) still works.
	bare := ui.ChatItem{IsSubagent: true, AgentType: model.AgentTypeGeneral, SubagentIdx: 0}
	if got := bare.WhoLabel(); got != "Agent" {
		t.Errorf("expected 'Agent', got %q", got)
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./internal/ui/... -run TestWhoLabel_TreeConnector -v
```

Expected: FAIL — `ChatItem` has no `TreeConnector` field.

**Step 3: Add `TreeConnector` field to `ChatItem`**

In `internal/ui/chat_item.go`, add after `IsDivider bool`:

```go
TreeConnector string // "├─", "└─", or "" for non-sub-agent items
```

**Step 4: Update `WhoLabel()` to prepend the connector**

Replace the sub-agent branch in `WhoLabel()`:

```go
if c.IsSubagent {
    name := agentDisplayName(c.AgentType)
    if c.TreeConnector != "" {
        return c.TreeConnector + " " + name
    }
    return name
}
```

**Step 5: Run test to verify it passes**

```bash
go test ./internal/ui/... -run TestWhoLabel_TreeConnector -v
```

Expected: PASS

**Step 6: Run all tests to confirm no regressions**

```bash
make test
```

Expected: all pass.

**Step 7: Commit**

```bash
git add internal/ui/chat_item.go internal/ui/chat_item_test.go
git commit -m "feat: add TreeConnector field to ChatItem and update WhoLabel"
```

---

### Task 2: Assign connectors in `interleaveSubagents`

**Files:**
- Modify: `internal/ui/chat_item.go`
- Test: `internal/ui/chat_item_test.go`

**Step 1: Write failing tests**

Add to `internal/ui/chat_item_test.go`:

```go
func TestBuildChatItems_TreeConnectors_Single(t *testing.T) {
	// One Agent call → sole sub-agent gets "└─".
	mainTurns := []model.Turn{
		{Role: "user", Text: "go"},
		{Role: "assistant", Text: "delegating", ToolCalls: []*model.ToolCall{{Name: "Agent"}}},
	}
	sub := []model.Turn{{Role: "assistant", Text: "done"}}
	items := BuildChatItems(mainTurns, [][]model.Turn{sub}, []model.AgentType{model.AgentTypeExplore})

	var subItems []ChatItem
	for _, it := range items {
		if it.IsSubagent {
			subItems = append(subItems, it)
		}
	}
	if len(subItems) != 1 {
		t.Fatalf("expected 1 sub-agent item, got %d", len(subItems))
	}
	if subItems[0].TreeConnector != "└─" {
		t.Errorf("expected '└─', got %q", subItems[0].TreeConnector)
	}
}

func TestBuildChatItems_TreeConnectors_Multiple(t *testing.T) {
	// Two Agent calls in one parent turn → "├─" then "└─".
	mainTurns := []model.Turn{
		{Role: "user", Text: "go"},
		{Role: "assistant", Text: "delegating", ToolCalls: []*model.ToolCall{
			{Name: "Agent"},
			{Name: "Agent"},
		}},
	}
	sub0 := []model.Turn{{Role: "assistant", Text: "first"}}
	sub1 := []model.Turn{{Role: "assistant", Text: "second"}}
	items := BuildChatItems(
		mainTurns,
		[][]model.Turn{sub0, sub1},
		[]model.AgentType{model.AgentTypeExplore, model.AgentTypePlan},
	)

	var subItems []ChatItem
	for _, it := range items {
		if it.IsSubagent {
			subItems = append(subItems, it)
		}
	}
	if len(subItems) != 2 {
		t.Fatalf("expected 2 sub-agent items, got %d", len(subItems))
	}
	if subItems[0].TreeConnector != "├─" {
		t.Errorf("expected '├─' for first, got %q", subItems[0].TreeConnector)
	}
	if subItems[1].TreeConnector != "└─" {
		t.Errorf("expected '└─' for second, got %q", subItems[1].TreeConnector)
	}
}

func TestBuildChatItems_TreeConnectors_SeparateTurns(t *testing.T) {
	// Two parent turns, one Agent call each → each sub-agent gets "└─".
	mainTurns := []model.Turn{
		{Role: "user", Text: "go"},
		{Role: "assistant", Text: "first batch", ToolCalls: []*model.ToolCall{{Name: "Agent"}}},
		{Role: "assistant", Text: "second batch", ToolCalls: []*model.ToolCall{{Name: "Agent"}}},
	}
	sub0 := []model.Turn{{Role: "assistant", Text: "a"}}
	sub1 := []model.Turn{{Role: "assistant", Text: "b"}}
	items := BuildChatItems(
		mainTurns,
		[][]model.Turn{sub0, sub1},
		[]model.AgentType{model.AgentTypeExplore, model.AgentTypePlan},
	)

	var subItems []ChatItem
	for _, it := range items {
		if it.IsSubagent {
			subItems = append(subItems, it)
		}
	}
	if len(subItems) != 2 {
		t.Fatalf("expected 2 sub-agent items, got %d", len(subItems))
	}
	if subItems[0].TreeConnector != "└─" {
		t.Errorf("expected '└─' for sub0, got %q", subItems[0].TreeConnector)
	}
	if subItems[1].TreeConnector != "└─" {
		t.Errorf("expected '└─' for sub1, got %q", subItems[1].TreeConnector)
	}
}
```

**Step 2: Run tests to verify they fail**

```bash
go test ./internal/ui/... -run "TestBuildChatItems_TreeConnectors" -v
```

Expected: FAIL — `TreeConnector` is always `""`.

**Step 3: Update `interleaveSubagents` to assign connectors**

In `BuildChatItems`, replace the `interleaveSubagents` closure with:

```go
interleaveSubagents := func(toolCalls []*model.ToolCall) {
    // Pre-count eligible agent calls in this batch.
    batchSize := 0
    for _, tc := range toolCalls {
        if (tc.Name == "Task" || tc.Name == "Agent") && subIdx+batchSize < len(subagentTurns) {
            batchSize++
        }
    }
    batchPos := 0
    for _, tc := range toolCalls {
        if (tc.Name == "Task" || tc.Name == "Agent") && subIdx < len(subagentTurns) {
            agentType := model.AgentTypeGeneral
            if subIdx < len(subagentTypes) {
                agentType = subagentTypes[subIdx]
            }
            connector := "├─"
            if batchPos == batchSize-1 {
                connector = "└─"
            }
            batchPos++

            // Collect assistant turns, preferring one with text or tool calls as primary.
            var allAssistant []model.Turn
            for _, st := range subagentTurns[subIdx] {
                if st.Role != "assistant" {
                    continue
                }
                allAssistant = append(allAssistant, st)
            }
            var first *model.Turn
            var extra []model.Turn
            for i := range allAssistant {
                if first == nil {
                    if allAssistant[i].Text != "" || len(allAssistant[i].ToolCalls) > 0 {
                        first = &allAssistant[i]
                    } else {
                        extra = append(extra, allAssistant[i])
                    }
                } else {
                    extra = append(extra, allAssistant[i])
                }
            }
            if first == nil && len(allAssistant) > 0 {
                first = &allAssistant[0]
                extra = allAssistant[1:]
            }
            if first != nil {
                items = append(items, ChatItem{
                    Turn:          *first,
                    ExtraTurns:    extra,
                    IsSubagent:    true,
                    AgentType:     agentType,
                    SubagentIdx:   subIdx,
                    TreeConnector: connector,
                })
            }
            subIdx++
        }
    }
}
```

**Step 4: Run tests to verify they pass**

```bash
go test ./internal/ui/... -run "TestBuildChatItems_TreeConnectors" -v
```

Expected: PASS

**Step 5: Run all tests**

```bash
make test
```

Expected: all pass.

**Step 6: Commit**

```bash
git add internal/ui/chat_item.go internal/ui/chat_item_test.go
git commit -m "feat: assign tree connectors to sub-agent ChatItems"
```

---

### Task 3: Widen the NAME column

**Files:**
- Modify: `internal/view/chat.go`

**Step 1: Update NAME column width**

In `internal/view/chat.go`, change:

```go
{Title: "NAME", Width: 10},
```

to:

```go
{Title: "NAME", Width: 14},
```

**Step 2: Run all tests**

```bash
make test
```

Expected: all pass.

**Step 3: Verify visually with demo**

```bash
go run . --demo
```

Navigate to a session's chat view. Confirm:
- `└─ Explorer` / `├─ Explorer` appear in the NAME column without truncation
- Regular `Claude` / `You` rows are unchanged

**Step 4: Commit**

```bash
git add internal/view/chat.go
git commit -m "feat: widen NAME column to 14 for tree connector prefix"
```

---

### Task 4: Pre-completion checks

**Step 1: Format, lint, test**

```bash
make fmt && make lint && make test
```

Expected: all pass, no lint warnings.

**Step 2: Run demo and smoke-test**

```bash
go run . --demo
```

Navigate to the session with sub-agents. Confirm tree connectors render correctly.
