package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/Curt-Park/claudeview/internal/config"
	"github.com/Curt-Park/claudeview/internal/demo"
	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/parallel"
	"github.com/Curt-Park/claudeview/internal/provider"
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
		dp = demo.NewProvider()
	} else {
		dp = provider.NewLive(config.ClaudeDir())
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

	// Key-based cursor for history view: survives expansion (which shifts numeric indices).
	// historyCursorKey is the ChatItemKey of the selected ChatItem (or parent when on a sub-row).
	// historyToolCallID identifies the specific ToolCallRow sub-row; "" means cursor is on the parent.
	historyCursorKey  string
	historyToolCallID string
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
			rm.pluginItems = rm.dp.GetPluginItems(rm.app.SelectedPlugin)
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

	// For the history view, also maintain a key-based cursor so that expansion
	// (which inserts sub-rows and shifts numeric indices) doesn't corrupt position.
	if rm.lastResource == model.ResourceHistory {
		if row := rm.app.Table.SelectedRow(); row != nil {
			switch v := row.Data.(type) {
			case ui.ChatItem:
				rm.historyCursorKey = ui.ChatItemKey(v)
				rm.historyToolCallID = ""
			case ui.ToolCallRow:
				rm.historyCursorKey = v.ChatItemKey
				// Use tool call ID to identify the sub-row; fall back to name.
				rm.historyToolCallID = v.ToolCall.ID
				if rm.historyToolCallID == "" {
					rm.historyToolCallID = v.ToolCall.Name
				}
			}
		}
	}

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

		// Resolve the base-table cursor index from the key-based anchor when
		// expansion is active, to avoid the expanded-index being clamped to last row.
		sel := cur.sel
		if rm.historyCursorKey != "" && len(rm.app.ExpandedItems) > 0 {
			for i, item := range rm.chatItems {
				if ui.ChatItemKey(item) == rm.historyCursorKey {
					sel = i
					break
				}
			}
		}

		rm.app.Table = rm.chatView.Sync(rm.chatItems, w, h, sel, cur.off, flt, false)
		rm.app.ApplyExpansion()
		// If cursor was on a tool call sub-row, restore to that specific sub-row.
		if rm.historyToolCallID != "" {
			for i, row := range rm.app.Table.FilteredRows() {
				tr, ok := row.Data.(ui.ToolCallRow)
				if !ok || tr.ChatItemKey != rm.historyCursorKey {
					continue
				}
				id := tr.ToolCall.ID
				if id == "" {
					id = tr.ToolCall.Name
				}
				if id == rm.historyToolCallID {
					rm.app.Table.Selected = i
					rm.app.Table.EnsureVisible()
					break
				}
			}
		}
		if rm.app.ChatFollow {
			rm.app.Table.GotoBottom()
		}
		rm.app.RefreshMenu()
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

	case ui.SyncViewMsg:
		rm.syncView()
		return rm, nil

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
				msg.pluginItems = dp.GetPluginItems(selectedPlugin)
			}
		case model.ResourceMemory:
			msg.memories = dp.GetMemories(projectHash)
		case model.ResourceHistory, model.ResourceHistoryDetail:
			// Re-scan sessions to detect newly created sessions in the slug group.
			freshSlug := refreshSlugGroup(dp, projectHash, sessionID, slugSessions)
			if len(freshSlug) > 1 {
				msg.slugGroupSessions = freshSlug
				type slugResult struct {
					turns    []model.Turn
					subTurns [][]model.Turn
					subTypes []model.AgentType
				}
				results := parallel.Map(freshSlug, func(s *model.Session) slugResult {
					turns := dp.GetTurns(s.FilePath)
					var subTurns [][]model.Turn
					if s.SubagentDir != "" {
						subInfos, _ := transcript.ScanSubagents(s.SubagentDir)
						subTurns = parallel.Map(subInfos, func(si transcript.SessionInfo) []model.Turn {
							return dp.GetTurns(si.FilePath)
						})
					}
					return slugResult{
						turns:    turns,
						subTurns: subTurns,
						subTypes: model.ExtractSubagentTypes(turns),
					}
				})
				for _, r := range results {
					msg.slugGroupTurns = append(msg.slugGroupTurns, r.turns)
					msg.slugGroupSubTurns = append(msg.slugGroupSubTurns, r.subTurns)
					msg.slugGroupSubTypes = append(msg.slugGroupSubTypes, r.subTypes)
				}
			} else {
				if sessionFilePath != "" {
					msg.turns = dp.GetTurns(sessionFilePath)
				}
				if subagentDir != "" {
					subInfos, _ := transcript.ScanSubagents(subagentDir)
					msg.subagentTurns = parallel.Map(subInfos, func(si transcript.SessionInfo) []model.Turn {
						return dp.GetTurns(si.FilePath)
					})
				}
				msg.subagentTypes = model.ExtractSubagentTypes(msg.turns)
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
