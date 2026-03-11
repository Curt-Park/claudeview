package provider

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/Curt-Park/claudeview/internal/config"
	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/parallel"
	"github.com/Curt-Park/claudeview/internal/stringutil"
	"github.com/Curt-Park/claudeview/internal/transcript"
	"github.com/Curt-Park/claudeview/internal/ui"
)

// Live implements ui.DataProvider by reading from the Claude data directory.
type Live struct {
	claudeDir      string
	currentProject string
	currentSession string
	aggCache       map[string]*transcript.SessionAggregates
	turnsCache     map[string]*transcript.TranscriptCache
	mu             sync.Mutex
}

// NewLive creates a new Live provider.
func NewLive(claudeDir string) ui.DataProvider {
	return &Live{
		claudeDir:  claudeDir,
		aggCache:   make(map[string]*transcript.SessionAggregates),
		turnsCache: make(map[string]*transcript.TranscriptCache),
	}
}

func (l *Live) GetProjects() []*model.Project {
	infos, err := transcript.ScanProjects(l.claudeDir)
	if err != nil {
		return []*model.Project{}
	}
	var projects []*model.Project
	for _, info := range infos {
		p := &model.Project{
			Hash:     info.Hash,
			Path:     info.Path,
			LastSeen: info.LastSeen,
		}
		results := parallel.Map(info.Sessions, func(si transcript.SessionInfo) *model.Session {
			return l.sessionFromInfo(si)
		})
		for _, s := range results {
			if s.Topic == "" {
				continue // skip empty sessions (no user messages yet)
			}
			p.Sessions = append(p.Sessions, s)
		}
		projects = append(projects, p)
	}
	return projects
}

func (l *Live) GetSessions(projectHash string) []*model.Session {
	if projectHash != "" {
		l.currentProject = projectHash
	}

	infos, err := transcript.ScanProjects(l.claudeDir)
	if err != nil {
		return []*model.Session{}
	}

	type sessionWork struct {
		si          transcript.SessionInfo
		projectHash string
	}
	var work []sessionWork
	for _, info := range infos {
		if l.currentProject != "" && info.Hash != l.currentProject {
			continue
		}
		for _, si := range info.Sessions {
			work = append(work, sessionWork{si: si, projectHash: info.Hash})
		}
	}

	results := parallel.Map(work, func(w sessionWork) *model.Session {
		s := l.sessionFromInfo(w.si)
		s.ProjectHash = w.projectHash
		return s
	})

	var sessions []*model.Session
	for _, s := range results {
		if s.Topic == "" {
			continue // skip empty sessions (no user messages yet)
		}
		sessions = append(sessions, s)
	}
	return model.GroupSessionsBySlug(sessions)
}

func (l *Live) GetAgents(sessionID string) []*model.Agent {
	if sessionID != "" {
		l.currentSession = sessionID
	}
	sessions := l.GetSessions(l.currentProject)
	if sessionID == "" {
		var all []*model.Agent
		for _, s := range sessions {
			all = append(all, parseAgentsFromSession(s)...)
		}
		return all
	}
	for _, s := range sessions {
		if s.ID == sessionID {
			return parseAgentsFromSession(s)
		}
	}
	return []*model.Agent{}
}

func (l *Live) GetPlugins(projectHash string) []*model.Plugin {
	installed, err := config.LoadInstalledPlugins(l.claudeDir)
	if err != nil {
		return []*model.Plugin{}
	}
	globalEnabled, _ := config.EnabledPlugins(l.claudeDir)

	var plugins []*model.Plugin
	for _, p := range installed {
		if projectHash == "" && p.Scope != "user" {
			continue
		}
		key := p.Name + "@" + p.Marketplace
		var isEnabled bool
		if p.ProjectPath != "" {
			pluginEnabled := config.ProjectEnabledPlugins(p.ProjectPath)
			isEnabled = pluginEnabled[key]
		} else {
			isEnabled = globalEnabled[key]
		}
		plugins = append(plugins, &model.Plugin{
			Name:         p.Name,
			Version:      p.Version,
			Marketplace:  p.Marketplace,
			Scope:        p.Scope,
			Enabled:      isEnabled,
			InstalledAt:  p.InstalledAt,
			CacheDir:     p.CacheDir,
			SkillCount:   model.CountSkills(p.CacheDir),
			CommandCount: model.CountCommands(p.CacheDir),
			HookCount:    model.CountHooks(p.CacheDir),
			AgentCount:   model.CountAgents(p.CacheDir),
			MCPCount:     model.CountMCPs(p.CacheDir),
		})
	}
	return plugins
}

func (l *Live) GetPluginItems(plugin *model.Plugin) []*model.PluginItem {
	return model.ListPluginItems(plugin.CacheDir)
}

func (l *Live) GetMemories(projectHash string) []*model.Memory {
	if projectHash == "" {
		return []*model.Memory{}
	}
	memDir := filepath.Join(l.claudeDir, "projects", projectHash, "memory")
	entries, err := os.ReadDir(memDir)
	if err != nil {
		return []*model.Memory{}
	}
	var memories []*model.Memory
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".md" {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		path := filepath.Join(memDir, e.Name())
		memories = append(memories, &model.Memory{
			Name:    e.Name(),
			Path:    path,
			Title:   stringutil.MdTitle(path),
			Size:    info.Size(),
			ModTime: info.ModTime(),
		})
	}
	return memories
}

func (l *Live) GetTurns(filePath string) []model.Turn {
	l.mu.Lock()
	cached := l.turnsCache[filePath]
	l.mu.Unlock()

	cache, err := transcript.ParseFileIncremental(filePath, cached)
	if err != nil {
		return nil
	}

	l.mu.Lock()
	l.turnsCache[filePath] = cache
	l.mu.Unlock()

	parsed := cache.Turns()
	turns := make([]model.Turn, 0, len(parsed))
	for _, t := range parsed {
		turn := model.Turn{
			Role:            t.Role,
			Text:            t.Text,
			Thinking:        t.Thinking,
			ModelName:       t.Model,
			InputTokens:     t.Usage.NewInputTokens(),
			CacheReadTokens: t.Usage.CacheReadInputTokens,
			OutputTokens:    t.Usage.OutputTokens,
			Timestamp:       t.Timestamp,
		}
		for _, tc := range t.ToolCalls {
			turn.ToolCalls = append(turn.ToolCalls, &model.ToolCall{
				ID:        tc.ID,
				Name:      tc.Name,
				Input:     tc.Input,
				Result:    tc.Result,
				IsError:   tc.IsError,
				Timestamp: tc.Timestamp,
				Duration:  tc.Duration,
			})
		}
		turns = append(turns, turn)
	}
	return turns
}

// sessionFromInfo creates a Session model from a transcript SessionInfo using incremental parsing.
func (l *Live) sessionFromInfo(si transcript.SessionInfo) *model.Session {
	s := &model.Session{
		ID:          si.ID,
		FilePath:    si.FilePath,
		SubagentDir: si.SubagentDir,
		ModTime:     si.ModTime,
	}

	l.mu.Lock()
	cached := l.aggCache[si.FilePath]
	l.mu.Unlock()

	agg, err := transcript.ParseAggregatesIncremental(si.FilePath, cached)
	if err != nil {
		return s
	}

	l.mu.Lock()
	l.aggCache[si.FilePath] = agg
	l.mu.Unlock()

	s.NumTurns = agg.NumTurns
	s.Topic = agg.Topic
	s.Branch = agg.Branch
	s.Slug = agg.Slug
	s.ToolCallCount = agg.TotalToolCalls
	s.AgentCount = 1 + transcript.CountSubagents(si.SubagentDir)
	if info, err := os.Stat(si.FilePath); err == nil {
		s.FileSize = info.Size()
	}

	s.TokensByModel = make(map[string]model.TokenCount, len(agg.TokensByModel))
	for m, u := range agg.TokensByModel {
		s.TokensByModel[m] = model.TokenCount{
			InputTokens:     u.InputTokens,
			CacheReadTokens: u.CacheReadInputTokens,
			OutputTokens:    u.OutputTokens,
		}
	}

	// Merge subagent token data into session totals
	if si.SubagentDir != "" {
		subInfos, _ := transcript.ScanSubagents(si.SubagentDir)
		if len(subInfos) > 0 {
			subAggs := parallel.Map(subInfos, func(sub transcript.SessionInfo) *transcript.SessionAggregates {
				l.mu.Lock()
				subCached := l.aggCache[sub.FilePath]
				l.mu.Unlock()

				subAgg, err := transcript.ParseAggregatesIncremental(sub.FilePath, subCached)
				if err != nil {
					return nil
				}

				l.mu.Lock()
				l.aggCache[sub.FilePath] = subAgg
				l.mu.Unlock()

				return subAgg
			})
			for _, subAgg := range subAggs {
				if subAgg == nil {
					continue
				}
				for m, u := range subAgg.TokensByModel {
					cur := s.TokensByModel[m]
					cur.InputTokens += u.InputTokens
					cur.CacheReadTokens += u.CacheReadInputTokens
					cur.OutputTokens += u.OutputTokens
					s.TokensByModel[m] = cur
				}
			}
		}
	}

	return s
}

// populateToolCalls fills agent.ToolCalls from a parsed transcript.
func populateToolCalls(agent *model.Agent, sessionID string, parsed *transcript.ParsedTranscript) {
	for _, turn := range parsed.Turns {
		for _, tc := range turn.ToolCalls {
			agent.ToolCalls = append(agent.ToolCalls, &model.ToolCall{
				ID:        tc.ID,
				SessionID: sessionID,
				AgentID:   agent.ID,
				Name:      tc.Name,
				Input:     tc.Input,
				Result:    tc.Result,
				IsError:   tc.IsError,
				Timestamp: turn.Timestamp,
			})
		}
	}
	if len(agent.ToolCalls) > 0 {
		last := agent.ToolCalls[len(agent.ToolCalls)-1]
		agent.LastActivity = last.Name + " " + last.InputSummary()
	}
}

// parseAgentsFromSession loads transcript and extracts agents.
func parseAgentsFromSession(s *model.Session) []*model.Agent {
	mainAgent := &model.Agent{
		ID:         "",
		SessionID:  s.ID,
		Type:       model.AgentTypeMain,
		Status:     model.StatusEnded,
		FilePath:   s.FilePath,
		IsSubagent: false,
	}

	// Parse main transcript for tool calls and subagent type extraction
	var subTypes []model.AgentType
	if parsed, err := transcript.ParseFile(s.FilePath); err == nil {
		populateToolCalls(mainAgent, s.ID, parsed)
		var calls []model.ToolCallInfo
		for _, t := range parsed.Turns {
			if t.Role != "assistant" {
				continue
			}
			for _, tc := range t.ToolCalls {
				calls = append(calls, model.ToolCallInfo{Name: tc.Name, Input: tc.Input})
			}
		}
		subTypes = model.ExtractAgentTypesFromCalls(calls)
	}

	agents := []*model.Agent{mainAgent}

	// Load subagents
	if s.SubagentDir != "" {
		subInfos, err := transcript.ScanSubagents(s.SubagentDir)
		if err == nil {
			type subWork struct {
				si        transcript.SessionInfo
				agentType model.AgentType
			}
			items := make([]subWork, len(subInfos))
			for i, si := range subInfos {
				agentType := model.AgentTypeGeneral
				if i < len(subTypes) {
					agentType = subTypes[i]
				}
				items[i] = subWork{si: si, agentType: agentType}
			}
			subAgents := parallel.Map(items, func(item subWork) *model.Agent {
				sub := &model.Agent{
					ID:         item.si.ID,
					SessionID:  s.ID,
					Type:       item.agentType,
					Status:     model.StatusDone,
					FilePath:   item.si.FilePath,
					IsSubagent: true,
					StartTime:  item.si.ModTime,
				}
				if subParsed, err := transcript.ParseFile(item.si.FilePath); err == nil {
					populateToolCalls(sub, s.ID, subParsed)
				}
				return sub
			})
			agents = append(agents, subAgents...)
		}
	}

	return agents
}
