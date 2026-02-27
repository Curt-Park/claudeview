package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
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
	projects   []*model.Project
	sessions   []*model.Session
	agents     []*model.Agent
	toolCalls  []*model.ToolCall
	tasks      []*model.Task
	plugins    []*model.Plugin
	mcpServers []*model.MCPServer
	resource   model.ResourceType
}

// rootModel wraps AppModel and manages actual resource data.
type rootModel struct {
	app        ui.AppModel
	dp         ui.DataProvider
	projects   []*model.Project
	sessions   []*model.Session
	agents     []*model.Agent
	toolCalls  []*model.ToolCall
	tasks      []*model.Task
	plugins    []*model.Plugin
	mcpServers []*model.MCPServer

	// Resource views (eagerly initialized in newRootModel)
	projectsView *view.ResourceView[*model.Project]
	sessionsView *view.ResourceView[*model.Session]
	agentsView   *view.ResourceView[*model.Agent]
	toolsView    *view.ResourceView[*model.ToolCall]
	tasksView    *view.ResourceView[*model.Task]
	pluginsView  *view.ResourceView[*model.Plugin]
	mcpView      *view.ResourceView[*model.MCPServer]

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
		toolsView:     view.NewToolsView(0, 0),
		tasksView:     view.NewTasksView(0, 0),
		pluginsView:   view.NewPluginsView(0, 0),
		mcpView:       view.NewMCPView(0, 0),
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
	case model.ResourceTools:
		rm.toolCalls = rm.dp.GetTools(rm.app.SelectedAgentID)
	case model.ResourceTasks:
		rm.tasks = rm.dp.GetTasks(rm.app.SelectedSessionID)
	case model.ResourcePlugins:
		rm.plugins = rm.dp.GetPlugins()
	case model.ResourceMCP:
		rm.mcpServers = rm.dp.GetMCPServers()
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
	case model.ResourceTools:
		rm.app.Table = rm.toolsView.Sync(rm.toolCalls, w, h, sel, off, flt, rm.app.SelectedAgentID == "")
	case model.ResourceTasks:
		rm.app.Table = rm.tasksView.Sync(rm.tasks, w, h, sel, off, flt, rm.app.SelectedSessionID == "")
	case model.ResourcePlugins:
		rm.app.Table = rm.pluginsView.Sync(rm.plugins, w, h, sel, off, flt, false)
	case model.ResourceMCP:
		rm.app.Table = rm.mcpView.Sync(rm.mcpServers, w, h, sel, off, flt, false)
	}
}

func (rm *rootModel) updateInfo() {
	rm.app.Info.Project = rm.app.SelectedProjectHash
	sess := rm.app.SelectedSessionID
	if len(sess) > 8 {
		sess = sess[:8]
	}
	rm.app.Info.Session = sess
	rm.app.Info.User = rm.userStr
	rm.app.Info.ClaudeVersion = rm.claudeVersion
	rm.app.Info.AppVersion = AppVersion
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
			case model.ResourceTools:
				rm.toolCalls = msg.toolCalls
			case model.ResourceTasks:
				rm.tasks = msg.tasks
			case model.ResourcePlugins:
				rm.plugins = msg.plugins
			case model.ResourceMCP:
				rm.mcpServers = msg.mcpServers
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
	agentID := rm.app.SelectedAgentID
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
		case model.ResourceTools:
			msg.toolCalls = dp.GetTools(agentID)
		case model.ResourceTasks:
			msg.tasks = dp.GetTasks(sessionID)
		case model.ResourcePlugins:
			msg.plugins = dp.GetPlugins()
		case model.ResourceMCP:
			msg.mcpServers = dp.GetMCPServers()
		}
		return msg
	}
}

func (rm *rootModel) View() string {
	return rm.app.View()
}

// --- Demo Data Provider ---

type demoDataProvider struct {
	projects   []*model.Project
	plugins    []*model.Plugin
	mcpServers []*model.MCPServer
}

func newDemoProvider() ui.DataProvider {
	projects := demo.GenerateProjects()
	return &demoDataProvider{
		projects:   projects,
		plugins:    demo.GeneratePlugins(),
		mcpServers: demo.GenerateMCPServers(),
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
func (d *demoDataProvider) GetTools(agentID string) []*model.ToolCall {
	if agentID == "" {
		var all []*model.ToolCall
		for _, p := range d.projects {
			for _, s := range p.Sessions {
				for _, a := range s.Agents {
					all = append(all, a.ToolCalls...)
				}
			}
		}
		return all
	}
	for _, p := range d.projects {
		for _, s := range p.Sessions {
			for _, a := range s.Agents {
				if a.ID == agentID {
					return a.ToolCalls
				}
			}
		}
	}
	return []*model.ToolCall{}
}
func (d *demoDataProvider) GetTasks(sessionID string) []*model.Task {
	if len(d.projects) > 0 && len(d.projects[0].Sessions) > 0 {
		sid := d.projects[0].Sessions[0].ID
		if sessionID != "" {
			sid = sessionID
		}
		return demo.GenerateTasks(sid)
	}
	return []*model.Task{}
}
func (d *demoDataProvider) GetPlugins() []*model.Plugin       { return d.plugins }
func (d *demoDataProvider) GetMCPServers() []*model.MCPServer { return d.mcpServers }

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

func (l *liveDataProvider) GetTools(agentID string) []*model.ToolCall {
	agents := l.GetAgents(l.currentSession)
	if agentID == "" {
		var all []*model.ToolCall
		for _, a := range agents {
			all = append(all, a.ToolCalls...)
		}
		return all
	}
	for _, a := range agents {
		if a.ID == agentID {
			return a.ToolCalls
		}
	}
	return []*model.ToolCall{}
}

func (l *liveDataProvider) GetTasks(sessionID string) []*model.Task {
	if sessionID == "" {
		sessionID = l.currentSession
	}
	entries, err := config.LoadTasks(l.claudeDir, sessionID)
	if err != nil {
		return []*model.Task{}
	}
	var tasks []*model.Task
	for _, e := range entries {
		tasks = append(tasks, &model.Task{
			ID:          e.ID,
			SessionID:   sessionID,
			Subject:     e.Subject,
			Description: e.Description,
			Status:      model.Status(e.Status),
			Owner:       e.Owner,
			BlockedBy:   e.BlockedBy,
			Blocks:      e.Blocks,
			ActiveForm:  e.ActiveForm,
		})
	}
	return tasks
}

func (l *liveDataProvider) GetPlugins() []*model.Plugin {
	installed, err := config.LoadInstalledPlugins(l.claudeDir)
	if err != nil {
		return []*model.Plugin{}
	}
	enabled, _ := config.EnabledPlugins(l.claudeDir)

	var plugins []*model.Plugin
	for _, p := range installed {
		isEnabled := p.Enabled
		if enabled != nil {
			isEnabled = enabled[p.Name]
		}
		cacheDir := config.PluginCacheDir(l.claudeDir, p.Marketplace, p.Name, p.Version)
		plugins = append(plugins, &model.Plugin{
			Name:         p.Name,
			Version:      p.Version,
			Marketplace:  p.Marketplace,
			Enabled:      isEnabled,
			InstalledAt:  p.InstalledAt,
			CacheDir:     cacheDir,
			SkillCount:   model.CountSkills(cacheDir),
			CommandCount: model.CountCommands(cacheDir),
			HookCount:    model.CountHooks(cacheDir),
		})
	}
	return plugins
}

func (l *liveDataProvider) GetMCPServers() []*model.MCPServer {
	settings, err := config.LoadSettings(l.claudeDir)
	if err != nil {
		return []*model.MCPServer{}
	}
	var servers []*model.MCPServer
	for name, s := range settings.MCPServers {
		transport := s.Type
		if transport == "" {
			transport = "stdio"
		}
		servers = append(servers, &model.MCPServer{
			Name:      name,
			Transport: transport,
			Command:   s.Command,
			Args:      s.Args,
			URL:       s.URL,
			Status:    model.StatusRunning,
		})
	}
	return servers
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
	s.DurationMS = agg.DurationMS
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
