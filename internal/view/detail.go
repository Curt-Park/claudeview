package view

import (
	"fmt"
	"strings"

	"github.com/Curt-Park/claudeview/internal/model"
)

// SessionDetailLines returns detail lines for a session.
func SessionDetailLines(s *model.Session) []string {
	lines := []string{
		fmt.Sprintf("  ID          : %s", s.ID),
		fmt.Sprintf("  Topic       : %s", s.Topic),
		fmt.Sprintf("  Agents      : %d", s.AgentCount),
		fmt.Sprintf("  Turns       : %d", s.NumTurns),
		fmt.Sprintf("  Tokens      : %s", s.TokenString()),
		fmt.Sprintf("  Duration    : %.1fs", float64(s.DurationMS)/1000),
		fmt.Sprintf("  Last Active : %s", s.LastActive()),
		fmt.Sprintf("  File        : %s", s.FilePath),
	}
	for m, tc := range s.TokensByModel {
		lines = append(lines, fmt.Sprintf("    %s: in=%d out=%d", m, tc.InputTokens, tc.OutputTokens))
	}
	if s.SubagentDir != "" {
		lines = append(lines, fmt.Sprintf("  Subagents   : %s", s.SubagentDir))
	}
	return lines
}

// AgentDetailLines returns detail lines for an agent.
func AgentDetailLines(a *model.Agent) []string {
	agentType := string(a.Type)
	isSubagent := "no"
	if a.IsSubagent {
		isSubagent = "yes"
	}
	lines := []string{
		fmt.Sprintf("  ID         : %s", a.ShortID()),
		fmt.Sprintf("  Full ID    : %s", a.ID),
		fmt.Sprintf("  Session    : %s", a.SessionID),
		fmt.Sprintf("  Type       : %s", agentType),
		fmt.Sprintf("  Subagent   : %s", isSubagent),
		fmt.Sprintf("  Status     : %s", a.Status),
		fmt.Sprintf("  Tools      : %d", len(a.ToolCalls)),
		fmt.Sprintf("  Last Act   : %s", a.LastActivity),
		fmt.Sprintf("  File       : %s", a.FilePath),
	}
	return lines
}

// TaskDetailLines returns detail lines for a task.
func TaskDetailLines(t *model.Task) []string {
	blockedBy := strings.Join(t.BlockedBy, ", ")
	if blockedBy == "" {
		blockedBy = "none"
	}
	blocks := strings.Join(t.Blocks, ", ")
	if blocks == "" {
		blocks = "none"
	}
	lines := []string{
		fmt.Sprintf("  ID         : %s", t.ID),
		fmt.Sprintf("  Session    : %s", t.SessionID),
		fmt.Sprintf("  Status     : %s %s", t.StatusIcon(), t.Status),
		fmt.Sprintf("  Subject    : %s", t.Subject),
		fmt.Sprintf("  Owner      : %s", t.Owner),
		fmt.Sprintf("  Blocked By : %s", blockedBy),
		fmt.Sprintf("  Blocks     : %s", blocks),
		"",
		"  Description:",
	}
	if t.Description != "" {
		for line := range strings.SplitSeq(t.Description, "\n") {
			lines = append(lines, "    "+line)
		}
	} else {
		lines = append(lines, "    (none)")
	}
	return lines
}

// PluginDetailLines returns detail lines for a plugin.
func PluginDetailLines(p *model.Plugin) []string {
	status := "disabled"
	if p.Enabled {
		status = "enabled"
	}
	lines := []string{
		fmt.Sprintf("  Name       : %s", p.Name),
		fmt.Sprintf("  Version    : %s", p.Version),
		fmt.Sprintf("  Marketplace: %s", p.Marketplace),
		fmt.Sprintf("  Status     : %s", status),
		fmt.Sprintf("  Skills     : %d", p.SkillCount),
		fmt.Sprintf("  Commands   : %d", p.CommandCount),
		fmt.Sprintf("  Hooks      : %d", p.HookCount),
		fmt.Sprintf("  Installed  : %s", p.InstalledAt),
		fmt.Sprintf("  Cache Dir  : %s", p.CacheDir),
	}
	return lines
}

// MCPDetailLines returns detail lines for an MCP server.
func MCPDetailLines(s *model.MCPServer) []string {
	args := strings.Join(s.Args, " ")
	lines := []string{
		fmt.Sprintf("  Name       : %s", s.Name),
		fmt.Sprintf("  Plugin     : %s", s.Plugin),
		fmt.Sprintf("  Transport  : %s", s.Transport),
		fmt.Sprintf("  Status     : %s", s.Status),
		fmt.Sprintf("  Tools      : %d", s.ToolCount),
		fmt.Sprintf("  Command    : %s", s.Command),
		fmt.Sprintf("  Args       : %s", args),
		fmt.Sprintf("  URL        : %s", s.URL),
	}
	return lines
}
