package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/Curt-Park/claudeview/internal/config"
	"github.com/Curt-Park/claudeview/internal/demo"
	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/transcript"
	"github.com/Curt-Park/claudeview/internal/ui"
	"github.com/Curt-Park/claudeview/internal/view"
)

var (
	demoMode     bool
	projectFlag  string
	resourceFlag string
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
	rootCmd.Flags().StringVar(&projectFlag, "project", "", "Filter to a specific project hash")
	rootCmd.Flags().StringVar(&resourceFlag, "resource", "sessions", "Initial resource to display (projects|sessions|agents|tools|tasks|plugins|mcp)")
}

func run(cmd *cobra.Command, args []string) error {
	initialResource := model.ResourceType(resourceFlag)
	if _, ok := model.ResolveResource(resourceFlag); !ok {
		initialResource = model.ResourceSessions
	}

	var dp ui.DataProvider
	if demoMode {
		dp = newDemoProvider()
	} else {
		claudeDir := config.ClaudeDir()
		dp = newLiveProvider(claudeDir, projectFlag)
	}

	appModel := ui.NewAppModel(dp, initialResource)

	// Pre-load initial data
	lp := dp.(*liveDataProvider)
	if demoMode {
		lp = nil
	}
	_ = lp

	// Create top-level model that wraps AppModel with actual view data
	root := newRootModel(appModel, dp)

	p := tea.NewProgram(root,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	_, err := p.Run()
	return err
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

	// Resource views
	projectsView *view.ProjectsView
	sessionsView *view.SessionsView
	agentsView   *view.AgentsView
	toolsView    *view.ToolsView
	tasksView    *view.TasksView
	pluginsView  *view.PluginsView
	mcpView      *view.MCPView
}

func newRootModel(app ui.AppModel, dp ui.DataProvider) *rootModel {
	rm := &rootModel{
		app: app,
		dp:  dp,
	}
	rm.loadData()
	return rm
}

func (rm *rootModel) Init() tea.Cmd {
	return tea.Batch(
		rm.app.Init(),
		rm.startWatcher(),
	)
}

type watcherMsg struct{}

func (rm *rootModel) startWatcher() tea.Cmd {
	if _, ok := rm.dp.(*demoDataProvider); ok {
		return nil
	}
	return func() tea.Msg {
		return watcherMsg{}
	}
}

func (rm *rootModel) loadData() {
	switch rm.app.Resource {
	case model.ResourceProjects:
		rm.loadProjects()
	case model.ResourceSessions:
		rm.loadSessions("")
	case model.ResourceAgents:
		rm.loadAgents("")
	case model.ResourceTools:
		rm.loadTools("")
	case model.ResourceTasks:
		rm.loadTasks("")
	case model.ResourcePlugins:
		rm.loadPlugins()
	case model.ResourceMCP:
		rm.loadMCP()
	}
	rm.syncView()
}

func (rm *rootModel) loadProjects() {
	raw := rm.dp.GetProjects()
	if projects, ok := raw.([]*model.Project); ok {
		rm.projects = projects
	}
}

func (rm *rootModel) loadSessions(projectHash string) {
	raw := rm.dp.GetSessions(projectHash)
	if sessions, ok := raw.([]*model.Session); ok {
		rm.sessions = sessions
	}
}

func (rm *rootModel) loadAgents(sessionID string) {
	raw := rm.dp.GetAgents(sessionID)
	if agents, ok := raw.([]*model.Agent); ok {
		rm.agents = agents
	}
}

func (rm *rootModel) loadTools(agentID string) {
	raw := rm.dp.GetTools(agentID)
	if tools, ok := raw.([]*model.ToolCall); ok {
		rm.toolCalls = tools
	}
}

func (rm *rootModel) loadTasks(sessionID string) {
	raw := rm.dp.GetTasks(sessionID)
	if tasks, ok := raw.([]*model.Task); ok {
		rm.tasks = tasks
	}
}

func (rm *rootModel) loadPlugins() {
	raw := rm.dp.GetPlugins()
	if plugins, ok := raw.([]*model.Plugin); ok {
		rm.plugins = plugins
	}
}

func (rm *rootModel) loadMCP() {
	raw := rm.dp.GetMCPServers()
	if servers, ok := raw.([]*model.MCPServer); ok {
		rm.mcpServers = servers
	}
}

func (rm *rootModel) syncView() {
	w := rm.app.Width
	h := rm.app.Height - 4
	if w == 0 {
		w = 120
	}
	if h <= 0 {
		h = 30
	}

	switch rm.app.Resource {
	case model.ResourceProjects:
		if rm.projectsView == nil {
			rm.projectsView = view.NewProjectsView(w, h)
		}
		rm.projectsView.Table.Width = w
		rm.projectsView.Table.Height = h
		rm.projectsView.SetProjects(rm.projects)
		rm.app.Table = rm.projectsView.Table
		rm.updateHeader()
	case model.ResourceSessions:
		if rm.sessionsView == nil {
			rm.sessionsView = view.NewSessionsView(w, h)
		}
		rm.sessionsView.Table.Width = w
		rm.sessionsView.Table.Height = h
		rm.sessionsView.SetSessions(rm.sessions)
		rm.app.Table = rm.sessionsView.Table
		rm.updateHeader()
	case model.ResourceAgents:
		if rm.agentsView == nil {
			rm.agentsView = view.NewAgentsView(w, h)
		}
		rm.agentsView.Table.Width = w
		rm.agentsView.Table.Height = h
		rm.agentsView.SetAgents(rm.agents)
		rm.app.Table = rm.agentsView.Table
	case model.ResourceTools:
		if rm.toolsView == nil {
			rm.toolsView = view.NewToolsView(w, h)
		}
		rm.toolsView.Table.Width = w
		rm.toolsView.Table.Height = h
		rm.toolsView.SetToolCalls(rm.toolCalls)
		rm.app.Table = rm.toolsView.Table
	case model.ResourceTasks:
		if rm.tasksView == nil {
			rm.tasksView = view.NewTasksView(w, h)
		}
		rm.tasksView.Table.Width = w
		rm.tasksView.Table.Height = h
		rm.tasksView.SetTasks(rm.tasks)
		rm.app.Table = rm.tasksView.Table
	case model.ResourcePlugins:
		if rm.pluginsView == nil {
			rm.pluginsView = view.NewPluginsView(w, h)
		}
		rm.pluginsView.Table.Width = w
		rm.pluginsView.Table.Height = h
		rm.pluginsView.SetPlugins(rm.plugins)
		rm.app.Table = rm.pluginsView.Table
	case model.ResourceMCP:
		if rm.mcpView == nil {
			rm.mcpView = view.NewMCPView(w, h)
		}
		rm.mcpView.Table.Width = w
		rm.mcpView.Table.Height = h
		rm.mcpView.SetServers(rm.mcpServers)
		rm.app.Table = rm.mcpView.Table
	}
}

func (rm *rootModel) updateHeader() {
	project := rm.dp.CurrentProject()
	rm.app.Header.ProjectName = project
	if lp, ok := rm.dp.(*liveDataProvider); ok {
		settings, _ := lp.getSettings()
		if settings != nil {
			rm.app.Header.Model = settings.Model
			rm.app.Header.MCPCount = len(settings.MCPServers)
		}
	}
}

func (rm *rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		rm.app.Width = msg.Width
		rm.app.Height = msg.Height
		rm.syncView()

	case ui.RefreshMsg:
		rm.loadData()
		rm.syncView()

	case watcherMsg:
		// start file watching in background
		go rm.watchFiles()
	}

	// Update app model
	newApp, cmd := rm.app.Update(msg)
	rm.app = newApp.(ui.AppModel)

	// Handle resource switch
	if rm.app.Resource != model.ResourceType("") {
		rm.syncViewAfterResourceChange()
	}

	return rm, cmd
}

func (rm *rootModel) syncViewAfterResourceChange() {
	// Detect if we need to reload data after navigation
	rm.loadData()
	rm.syncView()
}

func (rm *rootModel) watchFiles() {
	// File watching implementation would go here
	// This is a simplified version
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

func (d *demoDataProvider) GetProjects() any { return d.projects }
func (d *demoDataProvider) GetSessions(projectHash string) any {
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
func (d *demoDataProvider) GetAgents(sessionID string) any {
	for _, p := range d.projects {
		for _, s := range p.Sessions {
			if sessionID == "" || s.ID == sessionID {
				return s.Agents
			}
		}
	}
	return []*model.Agent{}
}
func (d *demoDataProvider) GetTools(agentID string) any {
	for _, p := range d.projects {
		for _, s := range p.Sessions {
			for _, a := range s.Agents {
				if agentID == "" || a.ID == agentID {
					return a.ToolCalls
				}
			}
		}
	}
	return []*model.ToolCall{}
}
func (d *demoDataProvider) GetTasks(sessionID string) any {
	if len(d.projects) > 0 && len(d.projects[0].Sessions) > 0 {
		sid := d.projects[0].Sessions[0].ID
		if sessionID != "" {
			sid = sessionID
		}
		return demo.GenerateTasks(sid)
	}
	return []*model.Task{}
}
func (d *demoDataProvider) GetPlugins() any    { return d.plugins }
func (d *demoDataProvider) GetMCPServers() any { return d.mcpServers }
func (d *demoDataProvider) CurrentProject() string {
	if len(d.projects) > 0 {
		return "my-awesome-app (demo)"
	}
	return ""
}
func (d *demoDataProvider) CurrentSession() string { return "" }
func (d *demoDataProvider) CurrentAgent() string   { return "" }

// --- Live Data Provider ---

type liveDataProvider struct {
	claudeDir      string
	projectFilter  string
	currentProject string
	currentSession string
	currentAgent   string
	cachedSettings *config.Settings
}

func newLiveProvider(claudeDir, projectFilter string) ui.DataProvider {
	return &liveDataProvider{
		claudeDir:     claudeDir,
		projectFilter: projectFilter,
	}
}

func (l *liveDataProvider) getSettings() (*config.Settings, error) {
	if l.cachedSettings != nil {
		return l.cachedSettings, nil
	}
	s, err := config.LoadSettings(l.claudeDir)
	if err == nil {
		l.cachedSettings = s
	}
	return s, err
}

func (l *liveDataProvider) GetProjects() any {
	infos, err := transcript.ScanProjects(l.claudeDir)
	if err != nil {
		return []*model.Project{}
	}
	var projects []*model.Project
	for _, info := range infos {
		if l.projectFilter != "" && !strings.Contains(info.Hash, l.projectFilter) {
			continue
		}
		p := &model.Project{
			Hash:     info.Hash,
			Path:     info.Path,
			LastSeen: info.LastSeen,
		}
		// Load sessions
		for _, si := range info.Sessions {
			s := sessionFromInfo(si)
			p.Sessions = append(p.Sessions, s)
		}
		projects = append(projects, p)
	}
	return projects
}

func (l *liveDataProvider) GetSessions(projectHash string) any {
	if projectHash != "" {
		l.currentProject = projectHash
	}

	infos, err := transcript.ScanProjects(l.claudeDir)
	if err != nil {
		return []*model.Session{}
	}

	var sessions []*model.Session
	for _, info := range infos {
		if l.projectFilter != "" && !strings.Contains(info.Hash, l.projectFilter) {
			continue
		}
		if l.currentProject != "" && info.Hash != l.currentProject {
			continue
		}
		for _, si := range info.Sessions {
			s := sessionFromInfo(si)
			s.ProjectHash = info.Hash
			sessions = append(sessions, s)
		}
	}
	return sessions
}

func (l *liveDataProvider) GetAgents(sessionID string) any {
	if sessionID != "" {
		l.currentSession = sessionID
	}
	sessions := l.GetSessions(l.currentProject).([]*model.Session)
	for _, s := range sessions {
		if sessionID == "" || s.ID == sessionID {
			// Parse transcript for this session
			agents := parseAgentsFromSession(s)
			return agents
		}
	}
	return []*model.Agent{}
}

func (l *liveDataProvider) GetTools(agentID string) any {
	if agentID != "" {
		l.currentAgent = agentID
	}
	agents := l.GetAgents(l.currentSession).([]*model.Agent)
	for _, a := range agents {
		if agentID == "" || a.ID == agentID {
			return a.ToolCalls
		}
	}
	return []*model.ToolCall{}
}

func (l *liveDataProvider) GetTasks(sessionID string) any {
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

func (l *liveDataProvider) GetPlugins() any {
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

func (l *liveDataProvider) GetMCPServers() any {
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

func (l *liveDataProvider) CurrentProject() string { return l.currentProject }
func (l *liveDataProvider) CurrentSession() string { return l.currentSession }
func (l *liveDataProvider) CurrentAgent() string   { return l.currentAgent }

// sessionFromInfo creates a Session model from a transcript SessionInfo.
func sessionFromInfo(si transcript.SessionInfo) *model.Session {
	s := &model.Session{
		ID:          si.ID,
		FilePath:    si.FilePath,
		SubagentDir: si.SubagentDir,
		ModTime:     si.ModTime,
		Status:      model.StatusEnded,
	}

	// Quick parse to get metadata
	parsed, err := transcript.ParseFile(si.FilePath)
	if err != nil {
		return s
	}

	s.TotalCost = parsed.TotalCost
	s.NumTurns = parsed.NumTurns
	s.DurationMS = parsed.DurationMS

	// Extract tokens and model from last assistant turn
	for i := len(parsed.Turns) - 1; i >= 0; i-- {
		t := parsed.Turns[i]
		if t.Role == "assistant" {
			s.InputTokens = t.Usage.InputTokens
			s.OutputTokens = t.Usage.OutputTokens
			s.Model = t.Model
			break
		}
	}

	// Determine status based on file modification vs now
	if time.Since(si.ModTime) < 30*time.Second {
		s.Status = model.StatusActive
	}

	return s
}

// parseAgentsFromSession loads transcript and extracts agents.
func parseAgentsFromSession(s *model.Session) []*model.Agent {
	mainAgent := &model.Agent{
		ID:         "",
		SessionID:  s.ID,
		Type:       model.AgentTypeMain,
		Status:     s.Status,
		FilePath:   s.FilePath,
		IsSubagent: false,
	}

	// Parse main transcript tool calls
	parsed, err := transcript.ParseFile(s.FilePath)
	if err == nil {
		for _, turn := range parsed.Turns {
			for _, tc := range turn.ToolCalls {
				mainAgent.ToolCalls = append(mainAgent.ToolCalls, &model.ToolCall{
					ID:        tc.ID,
					AgentID:   "",
					Name:      tc.Name,
					Input:     tc.Input,
					Result:    tc.Result,
					IsError:   tc.IsError,
					Timestamp: turn.Timestamp,
				})
			}
			if len(mainAgent.ToolCalls) > 0 {
				last := mainAgent.ToolCalls[len(mainAgent.ToolCalls)-1]
				mainAgent.LastActivity = last.Name + " " + last.InputSummary()
			}
		}
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
				subParsed, err := transcript.ParseFile(si.FilePath)
				if err == nil {
					for _, turn := range subParsed.Turns {
						for _, tc := range turn.ToolCalls {
							sub.ToolCalls = append(sub.ToolCalls, &model.ToolCall{
								ID:        tc.ID,
								AgentID:   si.ID,
								Name:      tc.Name,
								Input:     tc.Input,
								Result:    tc.Result,
								IsError:   tc.IsError,
								Timestamp: turn.Timestamp,
							})
						}
					}
					if len(sub.ToolCalls) > 0 {
						last := sub.ToolCalls[len(sub.ToolCalls)-1]
						sub.LastActivity = last.Name + " " + last.InputSummary()
					}
					// Detect agent type from subtype string in filename
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

// DurationMS is needed on Session model — add a field.
// We'll store it directly on Session.
func init() {
	// Register subcommands if needed in future
	_ = filepath.Join // ensure filepath is used
}
