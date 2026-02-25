package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/user"
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

// AppVersion is set from main.go via the build-time Version variable.
var AppVersion string

var (
	demoMode     bool
	projectFlag  string
	resourceFlag string
	renderOnce   bool
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
	rootCmd.Flags().StringVar(&resourceFlag, "resource", "projects", "Initial resource to display (projects|sessions|agents|tools|tasks|plugins|mcp)")
	rootCmd.Flags().BoolVar(&renderOnce, "render-once", false, "Render one frame to stdout and exit (for debugging)")
}

func run(cmd *cobra.Command, args []string) error {
	initialResource := model.ResourceType(resourceFlag)
	if _, ok := model.ResolveResource(resourceFlag); !ok {
		initialResource = model.ResourceProjects
	}

	var dp ui.DataProvider
	if demoMode {
		dp = newDemoProvider()
	} else {
		claudeDir := config.ClaudeDir()
		dp = newLiveProvider(claudeDir, projectFlag)
	}

	appModel := ui.NewAppModel(dp, initialResource)

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

	// Static info (set once at startup)
	userStr       string
	claudeVersion string
}

func newRootModel(app ui.AppModel, dp ui.DataProvider) *rootModel {
	rm := &rootModel{
		app:           app,
		dp:            dp,
		userStr:       currentUser(),
		claudeVersion: detectClaudeVersion(),
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
		rm.loadProjects()
	case model.ResourceSessions:
		rm.loadSessions(rm.app.SelectedProjectHash)
	case model.ResourceAgents:
		rm.loadAgents(rm.app.SelectedSessionID)
	case model.ResourceTools:
		rm.loadTools(rm.app.SelectedAgentID)
	case model.ResourceTasks:
		rm.loadTasks(rm.app.SelectedSessionID)
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
	h := rm.app.Height - 8
	if w <= 0 {
		w = 120
	}
	if h <= 0 {
		h = 30
	}

	rm.updateInfo()

	switch rm.app.Resource {
	case model.ResourceProjects:
		if rm.projectsView == nil {
			rm.projectsView = view.NewProjectsView(w, h)
		}
		rm.projectsView.Table.Width = w
		rm.projectsView.Table.Height = h
		rm.projectsView.SetProjects(rm.projects)
		rm.app.Table = rm.projectsView.Table
	case model.ResourceSessions:
		if rm.sessionsView == nil {
			rm.sessionsView = view.NewSessionsView(w, h)
		}
		rm.sessionsView.Table.Width = w
		rm.sessionsView.Table.Height = h
		rm.sessionsView.FlatMode = rm.app.SelectedProjectHash == ""
		rm.sessionsView.SetSessions(rm.sessions)
		rm.app.Table = rm.sessionsView.Table
	case model.ResourceAgents:
		if rm.agentsView == nil {
			rm.agentsView = view.NewAgentsView(w, h)
		}
		rm.agentsView.Table.Width = w
		rm.agentsView.Table.Height = h
		rm.agentsView.FlatMode = rm.app.SelectedSessionID == ""
		rm.agentsView.SetAgents(rm.agents)
		rm.app.Table = rm.agentsView.Table
	case model.ResourceTools:
		if rm.toolsView == nil {
			rm.toolsView = view.NewToolsView(w, h)
		}
		rm.toolsView.Table.Width = w
		rm.toolsView.Table.Height = h
		rm.toolsView.FlatMode = rm.app.SelectedAgentID == ""
		rm.toolsView.SetToolCalls(rm.toolCalls)
		rm.app.Table = rm.toolsView.Table
	case model.ResourceTasks:
		if rm.tasksView == nil {
			rm.tasksView = view.NewTasksView(w, h)
		}
		rm.tasksView.Table.Width = w
		rm.tasksView.Table.Height = h
		rm.tasksView.FlatMode = rm.app.SelectedSessionID == ""
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
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		rm.app.Width = msg.Width
		rm.app.Height = msg.Height
		rm.syncView()

	case ui.RefreshMsg:
		rm.loadData()
		rm.syncView()

	case ui.DetailRequestMsg:
		rm.populateDetail()

	case ui.LogRequestMsg:
		rm.populateLog()

	case ui.YAMLRequestMsg:
		rm.populateJSON()

	}

	// Update app model
	prevResource := rm.app.Resource
	newApp, cmd := rm.app.Update(msg)
	rm.app = newApp.(ui.AppModel)

	// Only reload data when resource changes
	if rm.app.Resource != prevResource {
		rm.loadData()
	}

	return rm, cmd
}

func (rm *rootModel) populateJSON() {
	row := rm.app.Table.SelectedRow()
	if row == nil {
		return
	}
	b, err := json.MarshalIndent(row.Data, "", "  ")
	if err != nil {
		rm.app.Detail.SetContentString(fmt.Sprintf("error: %v", err))
		return
	}
	rm.app.Detail.SetContentString(string(b))
}

func (rm *rootModel) populateDetail() {
	row := rm.app.Table.SelectedRow()
	if row == nil {
		return
	}
	var lines []string
	switch rm.app.Resource {
	case model.ResourceSessions:
		if s, ok := row.Data.(*model.Session); ok {
			lines = view.SessionDetailLines(s)
		}
	case model.ResourceAgents:
		if a, ok := row.Data.(*model.Agent); ok {
			lines = view.AgentDetailLines(a)
		}
	case model.ResourceTasks:
		if t, ok := row.Data.(*model.Task); ok {
			lines = view.TaskDetailLines(t)
		}
	case model.ResourcePlugins:
		if p, ok := row.Data.(*model.Plugin); ok {
			lines = view.PluginDetailLines(p)
		}
	case model.ResourceMCP:
		if s, ok := row.Data.(*model.MCPServer); ok {
			lines = view.MCPDetailLines(s)
		}
	default:
		// Fallback: JSON dump
		if b, err := json.MarshalIndent(row.Data, "  ", "  "); err == nil {
			lines = strings.Split(string(b), "\n")
		}
	}
	rm.app.Detail.SetContent(lines)
}

func (rm *rootModel) populateLog() {
	row := rm.app.Table.SelectedRow()
	if row == nil {
		return
	}

	var filePath string
	switch rm.app.Resource {
	case model.ResourceSessions:
		if s, ok := row.Data.(*model.Session); ok {
			filePath = s.FilePath
		}
	case model.ResourceAgents:
		if a, ok := row.Data.(*model.Agent); ok {
			filePath = a.FilePath
		}
	}

	if filePath == "" {
		rm.app.Log.SetLines([]ui.LogLine{{Text: "(no transcript file for this resource)", Style: "normal"}})
		return
	}

	parsed, err := transcript.ParseFile(filePath)
	if err != nil {
		rm.app.Log.SetLines([]ui.LogLine{{Text: fmt.Sprintf("error reading transcript: %v", err), Style: "normal"}})
		return
	}

	var logLines []ui.LogLine
	for _, turn := range parsed.Turns {
		ts := ""
		if !turn.Timestamp.IsZero() {
			ts = turn.Timestamp.Format("15:04:05")
		}
		header := fmt.Sprintf("[%s] %s", ts, turn.Role)
		logLines = append(logLines, ui.LogLine{Text: header, Style: "time"})

		if turn.Thinking != "" {
			for line := range strings.SplitSeq(turn.Thinking, "\n") {
				logLines = append(logLines, ui.LogLine{Text: "  <think> " + line, Style: "think"})
			}
		}
		if turn.Text != "" {
			for line := range strings.SplitSeq(turn.Text, "\n") {
				logLines = append(logLines, ui.LogLine{Text: "  " + line, Style: "text"})
			}
		}
		for _, tc := range turn.ToolCalls {
			logLines = append(logLines, ui.LogLine{Text: fmt.Sprintf("  ► %s", tc.Name), Style: "tool"})
			if len(tc.Input) > 0 && string(tc.Input) != "null" {
				input := string(tc.Input)
				if len(input) > 120 {
					input = input[:119] + "…"
				}
				logLines = append(logLines, ui.LogLine{Text: "    input: " + input, Style: "tool"})
			}
			if len(tc.Result) > 0 && string(tc.Result) != "null" {
				result := string(tc.Result)
				if len(result) > 120 {
					result = result[:119] + "…"
				}
				logLines = append(logLines, ui.LogLine{Text: "    result: " + result, Style: "result"})
			}
		}
	}

	if len(logLines) == 0 {
		logLines = []ui.LogLine{{Text: "(empty transcript)", Style: "normal"}}
	}
	rm.app.Log.SetLines(logLines)
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
func (d *demoDataProvider) GetTools(agentID string) any {
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

// --- Live Data Provider ---

type liveDataProvider struct {
	claudeDir      string
	projectFilter  string
	currentProject string
	currentSession string
	currentAgent   string
}

func newLiveProvider(claudeDir, projectFilter string) ui.DataProvider {
	return &liveDataProvider{
		claudeDir:     claudeDir,
		projectFilter: projectFilter,
	}
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
	if sessionID == "" {
		// Flat mode: return all agents from all sessions
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

func (l *liveDataProvider) GetTools(agentID string) any {
	if agentID != "" {
		l.currentAgent = agentID
	}
	agents := l.GetAgents(l.currentSession).([]*model.Agent)
	if agentID == "" {
		// Flat mode: return all tool calls from all agents
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

	// Determine status: active if JSONL was modified within the last 5 minutes
	// (covers pauses between turns while Claude Code is still running).
	if time.Since(si.ModTime) < 5*time.Minute {
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
					SessionID: s.ID,
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
								SessionID: s.ID,
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
