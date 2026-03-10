package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/Curt-Park/claudeview/internal/config"
	"github.com/Curt-Park/claudeview/internal/demo"
	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/transcript"
	"github.com/Curt-Park/claudeview/internal/ui"
	"github.com/Curt-Park/claudeview/internal/view"
)

// AppVersion is set from main.go via the build-time Version variable.
var AppVersion string

var demoMode bool

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "claudeview",
	Short: "Terminal dashboard for Claude Code sessions",
	Long: `claudeview is a terminal UI for monitoring Claude Code sessions,
tool calls, plugins, and MCP servers.

Navigate with j/k, Enter to drill down, / to filter, p for plugins, m for memories.`,
	RunE: run,
}

func init() {
	rootCmd.Flags().BoolVar(&demoMode, "demo", false, "Run with synthetic demo data")
}

func run(cmd *cobra.Command, args []string) error {
	var dp ui.DataProvider
	if demoMode {
		dp = newDemoProvider()
	} else {
		dp = newLiveProvider(config.ClaudeDir())
	}

	appModel := ui.NewAppModel(dp, model.ResourceProjects)

	// Create top-level model that wraps AppModel with actual view data
	root := newRootModel(appModel, dp)

	p := tea.NewProgram(root,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	_, err := p.Run()
	return err
}

// dataLoadedMsg carries freshly loaded data back to the UI goroutine.
type dataLoadedMsg struct {
	projects      []*model.Project
	sessions      []*model.Session
	plugins       []*model.Plugin
	pluginItems   []*model.PluginItem
	memories      []*model.Memory
	turns         []model.Turn
	subagentTurns [][]model.Turn
	subagentTypes []model.AgentType
	resource      model.ResourceType

	// Slug group reload data
	slugGroupSessions []*model.Session // refreshed slug group membership
	slugGroupTurns    [][]model.Turn
	slugGroupSubTurns [][][]model.Turn
	slugGroupSubTypes [][]model.AgentType
}

// rootModel wraps AppModel and manages actual resource data.
type rootModel struct {
	app         ui.AppModel
	dp          ui.DataProvider
	projects    []*model.Project
	sessions    []*model.Session
	plugins     []*model.Plugin
	pluginItems []*model.PluginItem
	memories    []*model.Memory

	// Resource views (eagerly initialized in newRootModel)
	projectsView    *view.ResourceView[*model.Project]
	sessionsView    *view.ResourceView[*model.Session]
	pluginsView     *view.ResourceView[*model.Plugin]
	pluginItemsView *view.ResourceView[*model.PluginItem]
	memoriesView    *view.ResourceView[*model.Memory]
	chatView        *view.ResourceView[ui.ChatItem]

	// Cached chat items for the chat table
	chatItems []ui.ChatItem

	// Static info (set once at startup)
	userStr       string
	claudeVersion string

	// Async loading state
	loading bool

	// Per-resource cursor state: each view remembers its own Selected/Offset.
	cursor       map[model.ResourceType]struct{ sel, off int }
	lastResource model.ResourceType
}

func newRootModel(app ui.AppModel, dp ui.DataProvider) *rootModel {
	rm := &rootModel{
		app:             app,
		dp:              dp,
		userStr:         currentUser(),
		claudeVersion:   detectClaudeVersion(),
		projectsView:    view.NewProjectsView(0, 0),
		sessionsView:    view.NewSessionsView(0, 0),
		pluginsView:     view.NewPluginsView(0, 0),
		pluginItemsView: view.NewPluginItemsView(0, 0),
		memoriesView:    view.NewMemoriesView(0, 0),
		chatView:        view.NewChatView(0, 0),
		cursor:          make(map[model.ResourceType]struct{ sel, off int }),
		lastResource:    app.Resource,
	}
	rm.loadData()
	return rm
}

func currentUser() string {
	if u, err := user.Current(); err == nil {
		return u.Username
	}
	return os.Getenv("USER")
}

func detectClaudeVersion() string {
	out, err := exec.Command("claude", "--version").Output()
	if err != nil {
		return "--"
	}
	v := strings.TrimSpace(string(out))
	// "claude X.Y.Z" or "X.Y.Z (Claude Code)" → keep only the version number part
	if parts := strings.Fields(v); len(parts) >= 1 {
		return parts[0]
	}
	return v
}

func (rm *rootModel) Init() tea.Cmd {
	return rm.app.Init()
}

func (rm *rootModel) loadData() {
	switch rm.app.Resource {
	case model.ResourceProjects:
		rm.projects = rm.dp.GetProjects()
	case model.ResourceSessions:
		rm.sessions = rm.dp.GetSessions(rm.app.SelectedProjectHash)
	case model.ResourcePlugins:
		rm.plugins = rm.dp.GetPlugins(rm.app.SelectedProjectHash)
	case model.ResourcePluginDetail:
		if rm.app.SelectedPlugin != nil {
			rm.pluginItems = model.ListPluginItems(rm.app.SelectedPlugin.CacheDir)
		}
	case model.ResourceMemory:
		rm.memories = rm.dp.GetMemories(rm.app.SelectedProjectHash)
	case model.ResourceHistory:
		rm.app.RebuildChatItems()
	}
	rm.syncView()
}

func (rm *rootModel) syncView() {
	w := rm.app.Width
	h := rm.app.ContentHeight()
	if w <= 0 {
		w = 120
	}
	if h <= 0 {
		h = 30
	}

	rm.updateInfo()

	// Save cursor position of the view we are leaving (or refreshing).
	rm.cursor[rm.lastResource] = struct{ sel, off int }{rm.app.Table.Selected, rm.app.Table.Offset}

	// Restore cursor for the resource now being displayed.
	rt := rm.app.Resource
	cur := rm.cursor[rt] // zero value {0, 0} on first visit
	flt := rm.app.Table.Filter

	switch rt {
	case model.ResourceProjects:
		rm.app.Table = rm.projectsView.Sync(rm.projects, w, h, cur.sel, cur.off, flt, false)
	case model.ResourceSessions:
		rm.app.Table = rm.sessionsView.Sync(rm.sessions, w, h, cur.sel, cur.off, flt, rm.app.SelectedProjectHash == "")
	case model.ResourcePlugins:
		rm.app.Table = rm.pluginsView.Sync(rm.plugins, w, h, cur.sel, cur.off, flt, false)
	case model.ResourcePluginDetail:
		rm.app.Table = rm.pluginItemsView.Sync(rm.pluginItems, w, h, cur.sel, cur.off, flt, false)
	case model.ResourceMemory:
		rm.app.Table = rm.memoriesView.Sync(rm.memories, w, h, cur.sel, cur.off, flt, false)
	case model.ResourceHistory:
		rm.chatItems = rm.app.ChatItems
		rm.app.Table = rm.chatView.Sync(rm.chatItems, w, h, cur.sel, cur.off, flt, false)
		if rm.app.ChatFollow {
			rm.app.Table.GotoBottom()
		}
	}

	rm.lastResource = rt
}

func (rm *rootModel) updateInfo() {
	rm.app.Info.Project = rm.app.SelectedProjectHash
	rm.app.Info.Session = rm.app.SelectedSessionSlug
	rm.app.Info.User = rm.userStr
	rm.app.Info.ClaudeVersion = rm.claudeVersion
	rm.app.Info.AppVersion = AppVersion
	rm.app.Info.MemoriesActive = rm.app.SelectedProjectHash != ""
	rm.app.Info.Resource = rm.app.Resource
}

func (rm *rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var extraCmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		rm.app.Width = msg.Width
		rm.app.Height = msg.Height
		rm.syncView()

	case ui.TickMsg:
		// Refresh time-based columns (LAST ACTIVE) on every tick
		rm.syncView()
		// Trigger async data reload if not already in progress
		if !rm.loading {
			rm.loading = true
			extraCmd = rm.loadDataAsync()
		}

	case dataLoadedMsg:
		rm.loading = false
		if msg.resource == rm.app.Resource {
			switch msg.resource {
			case model.ResourceProjects:
				rm.projects = msg.projects
			case model.ResourceSessions:
				rm.sessions = msg.sessions
			case model.ResourcePlugins:
				rm.plugins = msg.plugins
			case model.ResourcePluginDetail:
				rm.pluginItems = msg.pluginItems
			case model.ResourceMemory:
				rm.memories = msg.memories
			case model.ResourceHistory, model.ResourceHistoryDetail:
				// Update slug group membership if new sessions were detected.
				if len(msg.slugGroupSessions) > 0 {
					rm.app.SlugSessions = msg.slugGroupSessions
				}
				if len(msg.slugGroupTurns) > 1 {
					rm.app.SetSlugGroupData(msg.slugGroupTurns, msg.slugGroupSubTurns, msg.slugGroupSubTypes)
				} else {
					rm.app.SelectedTurns = msg.turns
					rm.app.SubagentTurns = msg.subagentTurns
					rm.app.SubagentTypes = msg.subagentTypes
				}
				rm.app.RebuildChatItems()
				rm.chatItems = rm.app.ChatItems
			}
			rm.syncView()
		}

	}

	// Update app model — must always run so AppModel can re-schedule tick()
	prevResource := rm.app.Resource
	newApp, cmd := rm.app.Update(msg)
	rm.app = newApp.(ui.AppModel)

	// Only reload data when resource changes
	if rm.app.Resource != prevResource {
		rm.loadData()
	}

	if extraCmd != nil {
		return rm, tea.Batch(cmd, extraCmd)
	}
	return rm, cmd
}

// loadDataAsync returns a tea.Cmd that loads data in a background goroutine.
func (rm *rootModel) loadDataAsync() tea.Cmd {
	resource := rm.app.Resource
	projectHash := rm.app.SelectedProjectHash
	sessionID := rm.app.SelectedSessionID
	selectedPlugin := rm.app.SelectedPlugin
	sessionFilePath := rm.app.SelectedSessionFilePath
	subagentDir := rm.app.SelectedSessionSubagentDir
	slugSessions := rm.app.SlugSessions
	dp := rm.dp
	return func() tea.Msg {
		msg := dataLoadedMsg{resource: resource}
		switch resource {
		case model.ResourceProjects:
			msg.projects = dp.GetProjects()
		case model.ResourceSessions:
			msg.sessions = dp.GetSessions(projectHash)
		case model.ResourcePlugins:
			msg.plugins = dp.GetPlugins(projectHash)
		case model.ResourcePluginDetail:
			if selectedPlugin != nil {
				msg.pluginItems = model.ListPluginItems(selectedPlugin.CacheDir)
			}
		case model.ResourceMemory:
			msg.memories = dp.GetMemories(projectHash)
		case model.ResourceHistory, model.ResourceHistoryDetail:
			// Re-scan sessions to detect newly created sessions in the slug group.
			freshSlug := refreshSlugGroup(dp, projectHash, sessionID, slugSessions)
			if len(freshSlug) > 1 {
				msg.slugGroupSessions = freshSlug
				for _, s := range freshSlug {
					turns := dp.GetTurns(s.FilePath)
					msg.slugGroupTurns = append(msg.slugGroupTurns, turns)
					var subTurns [][]model.Turn
					if s.SubagentDir != "" {
						subInfos, _ := transcript.ScanSubagents(s.SubagentDir)
						for _, si := range subInfos {
							subTurns = append(subTurns, dp.GetTurns(si.FilePath))
						}
					}
					msg.slugGroupSubTurns = append(msg.slugGroupSubTurns, subTurns)
					msg.slugGroupSubTypes = append(msg.slugGroupSubTypes, extractSubagentTypes(turns))
				}
			} else {
				if sessionFilePath != "" {
					msg.turns = dp.GetTurns(sessionFilePath)
				}
				if subagentDir != "" {
					subInfos, _ := transcript.ScanSubagents(subagentDir)
					for _, si := range subInfos {
						msg.subagentTurns = append(msg.subagentTurns, dp.GetTurns(si.FilePath))
					}
				}
				msg.subagentTypes = extractSubagentTypes(msg.turns)
			}
		}
		return msg
	}
}

// refreshSlugGroup re-scans sessions for the project and returns the updated
// slug group membership. This detects newly created sessions under the same slug.
// Returns the fresh slug group if it has >1 members, otherwise nil (single session).
func refreshSlugGroup(dp ui.DataProvider, projectHash, sessionID string, currentSlug []*model.Session) []*model.Session {
	// Determine the slug we're tracking.
	var slug string
	if len(currentSlug) > 0 {
		slug = currentSlug[0].Slug
	}
	if slug == "" {
		// Not a slug group — check if a new session joined to form one.
		allSessions := dp.GetSessions(projectHash)
		for _, s := range allSessions {
			if s.ID == sessionID && s.IsGroupRepresentative() {
				return s.GroupSessions
			}
		}
		return nil
	}
	// Re-scan and find the matching slug group.
	allSessions := dp.GetSessions(projectHash)
	for _, s := range allSessions {
		if s.Slug == slug && s.IsGroupRepresentative() {
			return s.GroupSessions
		}
	}
	return currentSlug
}

func (rm *rootModel) View() string {
	return rm.app.View()
}

// --- Demo Data Provider ---

type demoDataProvider struct {
	projects []*model.Project
	plugins  []*model.Plugin
}

func newDemoProvider() ui.DataProvider {
	return &demoDataProvider{
		projects: demo.GenerateProjects(),
		plugins:  demo.GeneratePlugins(),
	}
}

func (d *demoDataProvider) GetProjects() []*model.Project { return d.projects }
func (d *demoDataProvider) GetSessions(projectHash string) []*model.Session {
	for _, p := range d.projects {
		if projectHash == "" || p.Hash == projectHash {
			return p.Sessions
		}
	}
	if len(d.projects) > 0 {
		return d.projects[0].Sessions
	}
	return []*model.Session{}
}
func (d *demoDataProvider) GetAgents(sessionID string) []*model.Agent {
	if sessionID == "" {
		var all []*model.Agent
		for _, p := range d.projects {
			for _, s := range p.Sessions {
				all = append(all, s.Agents...)
			}
		}
		return all
	}
	for _, p := range d.projects {
		for _, s := range p.Sessions {
			if s.ID == sessionID {
				return s.Agents
			}
		}
	}
	return []*model.Agent{}
}
func (d *demoDataProvider) GetPlugins(_ string) []*model.Plugin { return d.plugins }
func (d *demoDataProvider) GetMemories(_ string) []*model.Memory {
	return demo.GenerateMemories()
}
func (d *demoDataProvider) GetTurns(_ string) []model.Turn { return nil }

// --- Live Data Provider ---

type liveDataProvider struct {
	claudeDir      string
	currentProject string
	currentSession string
	aggCache       map[string]*transcript.SessionAggregates
	turnsCache     map[string]*transcript.TranscriptCache
	mu             sync.Mutex
}

func newLiveProvider(claudeDir string) ui.DataProvider {
	return &liveDataProvider{
		claudeDir:  claudeDir,
		aggCache:   make(map[string]*transcript.SessionAggregates),
		turnsCache: make(map[string]*transcript.TranscriptCache),
	}
}

func (l *liveDataProvider) GetProjects() []*model.Project {
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
		for _, si := range info.Sessions {
			s := l.sessionFromInfo(si)
			if s.Topic == "" {
				continue // skip empty sessions (no user messages yet)
			}
			p.Sessions = append(p.Sessions, s)
		}
		projects = append(projects, p)
	}
	return projects
}

func (l *liveDataProvider) GetSessions(projectHash string) []*model.Session {
	if projectHash != "" {
		l.currentProject = projectHash
	}

	infos, err := transcript.ScanProjects(l.claudeDir)
	if err != nil {
		return []*model.Session{}
	}

	var sessions []*model.Session
	for _, info := range infos {
		if l.currentProject != "" && info.Hash != l.currentProject {
			continue
		}
		for _, si := range info.Sessions {
			s := l.sessionFromInfo(si)
			if s.Topic == "" {
				continue // skip empty sessions (no user messages yet)
			}
			s.ProjectHash = info.Hash
			sessions = append(sessions, s)
		}
	}
	return model.GroupSessionsBySlug(sessions)
}

func (l *liveDataProvider) GetAgents(sessionID string) []*model.Agent {
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

func (l *liveDataProvider) GetPlugins(projectHash string) []*model.Plugin {
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

func (l *liveDataProvider) GetMemories(projectHash string) []*model.Memory {
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
			Title:   mdTitle(path),
			Size:    info.Size(),
			ModTime: info.ModTime(),
		})
	}
	return memories
}

// sessionFromInfo creates a Session model from a transcript SessionInfo using incremental parsing.
func (l *liveDataProvider) sessionFromInfo(si transcript.SessionInfo) *model.Session {
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
		s.TokensByModel[m] = model.TokenCount{InputTokens: u.InputTokens, OutputTokens: u.OutputTokens}
	}

	// Merge subagent token data into session totals
	if si.SubagentDir != "" {
		subInfos, _ := transcript.ScanSubagents(si.SubagentDir)
		for _, sub := range subInfos {
			l.mu.Lock()
			subCached := l.aggCache[sub.FilePath]
			l.mu.Unlock()

			subAgg, err := transcript.ParseAggregatesIncremental(sub.FilePath, subCached)
			if err != nil {
				continue
			}

			l.mu.Lock()
			l.aggCache[sub.FilePath] = subAgg
			l.mu.Unlock()

			for m, u := range subAgg.TokensByModel {
				cur := s.TokensByModel[m]
				cur.InputTokens += u.InputTokens
				cur.OutputTokens += u.OutputTokens
				s.TokensByModel[m] = cur
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
		var calls []toolCallInfo
		for _, t := range parsed.Turns {
			if t.Role != "assistant" {
				continue
			}
			for _, tc := range t.ToolCalls {
				calls = append(calls, toolCallInfo{Name: tc.Name, Input: tc.Input})
			}
		}
		subTypes = extractAgentTypesFromCalls(calls)
	}

	agents := []*model.Agent{mainAgent}

	// Load subagents
	if s.SubagentDir != "" {
		subInfos, err := transcript.ScanSubagents(s.SubagentDir)
		if err == nil {
			for i, si := range subInfos {
				agentType := model.AgentTypeGeneral
				if i < len(subTypes) {
					agentType = subTypes[i]
				}
				sub := &model.Agent{
					ID:         si.ID,
					SessionID:  s.ID,
					Type:       agentType,
					Status:     model.StatusDone,
					FilePath:   si.FilePath,
					IsSubagent: true,
					StartTime:  si.ModTime,
				}
				// Parse subagent transcript
				if subParsed, err := transcript.ParseFile(si.FilePath); err == nil {
					populateToolCalls(sub, s.ID, subParsed)
				}
				agents = append(agents, sub)
			}
		}
	}

	return agents
}

func (l *liveDataProvider) GetTurns(filePath string) []model.Turn {
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
			Role:         t.Role,
			Text:         t.Text,
			Thinking:     t.Thinking,
			ModelName:    t.Model,
			InputTokens:  t.Usage.InputTokens,
			OutputTokens: t.Usage.OutputTokens,
			Timestamp:    t.Timestamp,
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

// mdTitle reads the first `# Heading` line from a markdown file and returns
// the heading text. Returns an empty string if the file cannot be read or
// contains no top-level heading.
func mdTitle(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	for _, line := range strings.SplitAfter(string(data), "\n") {
		line = strings.TrimRight(line, "\r\n")
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
	}
	return ""
}

// toolCallInfo holds the minimal fields needed to extract subagent types,
// allowing a single implementation to work with both model.Turn and transcript.Turn.
type toolCallInfo struct {
	Name  string
	Input json.RawMessage
}

// extractAgentTypesFromCalls reads Agent/Task tool calls and returns the subagent_type
// value for each, in call order. This matches the positional order used by
// BuildChatItems to interleave subagent turns.
func extractAgentTypesFromCalls(calls []toolCallInfo) []model.AgentType {
	var types []model.AgentType
	for _, c := range calls {
		if c.Name != "Agent" && c.Name != "Task" {
			continue
		}
		types = append(types, model.AgentTypeFromInput(c.Input))
	}
	return types
}

// extractSubagentTypes collects Agent/Task tool calls from model.Turn slices
// and delegates to extractAgentTypesFromCalls.
func extractSubagentTypes(turns []model.Turn) []model.AgentType {
	var calls []toolCallInfo
	for _, t := range turns {
		if t.Role != "assistant" {
			continue
		}
		for _, tc := range t.ToolCalls {
			calls = append(calls, toolCallInfo{Name: tc.Name, Input: tc.Input})
		}
	}
	return extractAgentTypesFromCalls(calls)
}
