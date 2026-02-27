package cmd

import (
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

var (
	demoMode   bool
	renderOnce bool
)

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
agents, tool calls, tasks, plugins, and MCP servers.

Navigate with j/k, Enter to drill down, l for logs, d for detail.
Use : to switch resource types (sessions, agents, tools, tasks, plugins, mcp).`,
	RunE: run,
}

func init() {
	rootCmd.Flags().BoolVar(&demoMode, "demo", false, "Run with synthetic demo data")
	rootCmd.Flags().BoolVar(&renderOnce, "render-once", false, "Render one frame to stdout and exit (for debugging)")
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

	if renderOnce {
		root.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
		fmt.Print(root.View())
		return nil
	}

	p := tea.NewProgram(root,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	_, err := p.Run()
	return err
}

// dataLoadedMsg carries freshly loaded data back to the UI goroutine.
type dataLoadedMsg struct {
	projects []*model.Project
	sessions []*model.Session
	agents   []*model.Agent
	plugins  []*model.Plugin
	memories []*model.Memory
	resource model.ResourceType
}

// rootModel wraps AppModel and manages actual resource data.
type rootModel struct {
	app      ui.AppModel
	dp       ui.DataProvider
	projects []*model.Project
	sessions []*model.Session
	agents   []*model.Agent
	plugins  []*model.Plugin
	memories []*model.Memory

	// Resource views (eagerly initialized in newRootModel)
	projectsView *view.ResourceView[*model.Project]
	sessionsView *view.ResourceView[*model.Session]
	agentsView   *view.ResourceView[*model.Agent]
	pluginsView  *view.ResourceView[*model.Plugin]
	memoriesView *view.ResourceView[*model.Memory]

	// Static info (set once at startup)
	userStr       string
	claudeVersion string

	// Async loading state
	loading bool
}

func newRootModel(app ui.AppModel, dp ui.DataProvider) *rootModel {
	rm := &rootModel{
		app:           app,
		dp:            dp,
		userStr:       currentUser(),
		claudeVersion: detectClaudeVersion(),
		projectsView:  view.NewProjectsView(0, 0),
		sessionsView:  view.NewSessionsView(0, 0),
		agentsView:    view.NewAgentsView(0, 0),
		pluginsView:   view.NewPluginsView(0, 0),
		memoriesView:  view.NewMemoriesView(0, 0),
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
	case model.ResourceAgents:
		rm.agents = rm.dp.GetAgents(rm.app.SelectedSessionID)
	case model.ResourcePlugins:
		rm.plugins = rm.dp.GetPlugins()
	case model.ResourceMemory:
		rm.memories = rm.dp.GetMemories(rm.app.SelectedProjectHash)
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

	// Preserve the user's cursor/scroll/filter state across rebuilds so that
	// RefreshMsg and WindowSizeMsg don't reset the selection to row 0.
	sel := rm.app.Table.Selected
	off := rm.app.Table.Offset
	flt := rm.app.Table.Filter

	switch rm.app.Resource {
	case model.ResourceProjects:
		rm.app.Table = rm.projectsView.Sync(rm.projects, w, h, sel, off, flt, false)
	case model.ResourceSessions:
		rm.app.Table = rm.sessionsView.Sync(rm.sessions, w, h, sel, off, flt, rm.app.SelectedProjectHash == "")
	case model.ResourceAgents:
		rm.app.Table = rm.agentsView.Sync(rm.agents, w, h, sel, off, flt, rm.app.SelectedSessionID == "")
	case model.ResourcePlugins:
		rm.app.Table = rm.pluginsView.Sync(rm.plugins, w, h, sel, off, flt, false)
	case model.ResourceMemory:
		rm.app.Table = rm.memoriesView.Sync(rm.memories, w, h, sel, off, flt, false)
	}
}

func (rm *rootModel) updateInfo() {
	rm.app.Info.Project = rm.app.SelectedProjectHash
	rm.app.Info.Session = view.ShortID(rm.app.SelectedSessionID, 8)
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
			case model.ResourceAgents:
				rm.agents = msg.agents
			case model.ResourcePlugins:
				rm.plugins = msg.plugins
			case model.ResourceMemory:
				rm.memories = msg.memories
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
	dp := rm.dp
	return func() tea.Msg {
		msg := dataLoadedMsg{resource: resource}
		switch resource {
		case model.ResourceProjects:
			msg.projects = dp.GetProjects()
		case model.ResourceSessions:
			msg.sessions = dp.GetSessions(projectHash)
		case model.ResourceAgents:
			msg.agents = dp.GetAgents(sessionID)
		case model.ResourcePlugins:
			msg.plugins = dp.GetPlugins()
		case model.ResourceMemory:
			msg.memories = dp.GetMemories(projectHash)
		}
		return msg
	}
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
func (d *demoDataProvider) GetPlugins() []*model.Plugin { return d.plugins }
func (d *demoDataProvider) GetMemories(_ string) []*model.Memory {
	return demo.GenerateMemories()
}

// --- Live Data Provider ---

type liveDataProvider struct {
	claudeDir      string
	currentProject string
	currentSession string
	aggCache       map[string]*transcript.SessionAggregates
	mu             sync.Mutex
}

func newLiveProvider(claudeDir string) ui.DataProvider {
	return &liveDataProvider{
		claudeDir: claudeDir,
		aggCache:  make(map[string]*transcript.SessionAggregates),
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
	return sessions
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

func (l *liveDataProvider) GetPlugins() []*model.Plugin {
	installed, err := config.LoadInstalledPlugins(l.claudeDir)
	if err != nil {
		return []*model.Plugin{}
	}
	enabled, _ := config.EnabledPlugins(l.claudeDir)

	var plugins []*model.Plugin
	for _, p := range installed {
		key := p.Name + "@" + p.Marketplace
		isEnabled := enabled[key]
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
	s.ToolCallCount = agg.TotalToolCalls
	s.AgentCount = 1 + transcript.CountSubagents(si.SubagentDir)
	if info, err := os.Stat(si.FilePath); err == nil {
		s.FileSize = info.Size()
	}

	s.TokensByModel = make(map[string]model.TokenCount, len(agg.TokensByModel))
	for m, u := range agg.TokensByModel {
		s.TokensByModel[m] = model.TokenCount{InputTokens: u.InputTokens, OutputTokens: u.OutputTokens}
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

	// Parse main transcript tool calls
	if parsed, err := transcript.ParseFile(s.FilePath); err == nil {
		populateToolCalls(mainAgent, s.ID, parsed)
	}

	agents := []*model.Agent{mainAgent}

	// Load subagents
	if s.SubagentDir != "" {
		subInfos, err := transcript.ScanSubagents(s.SubagentDir)
		if err == nil {
			for _, si := range subInfos {
				sub := &model.Agent{
					ID:         si.ID,
					SessionID:  s.ID,
					Type:       model.AgentTypeGeneral,
					Status:     model.StatusDone,
					FilePath:   si.FilePath,
					IsSubagent: true,
					StartTime:  si.ModTime,
				}
				// Parse subagent transcript
				if subParsed, err := transcript.ParseFile(si.FilePath); err == nil {
					populateToolCalls(sub, s.ID, subParsed)
					sub.Type = detectAgentType(si.ID)
				}
				agents = append(agents, sub)
			}
		}
	}

	return agents
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

// detectAgentType infers agent type from its ID or filename.
func detectAgentType(id string) model.AgentType {
	lower := strings.ToLower(id)
	switch {
	case strings.Contains(lower, "explore"):
		return model.AgentTypeExplore
	case strings.Contains(lower, "plan"):
		return model.AgentTypePlan
	case strings.Contains(lower, "bash"):
		return model.AgentTypeBash
	default:
		return model.AgentTypeGeneral
	}
}
