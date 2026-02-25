package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Curt-Park/claudeview/internal/model"
)

// ViewMode indicates what the content area is showing.
type ViewMode int

const (
	ModeTable  ViewMode = iota
	ModeLog             // l key
	ModeDetail          // d key
	ModeYAML            // y key — JSON dump of selected row
)

// TickMsg is sent on each timer tick for animations.
type TickMsg time.Time

// RefreshMsg signals data has been refreshed.
type RefreshMsg struct{}

// DetailRequestMsg signals that the detail view should be populated.
type DetailRequestMsg struct{}

// LogRequestMsg signals that the log view should be populated.
type LogRequestMsg struct{}

// YAMLRequestMsg signals that the JSON dump view should be populated.
type YAMLRequestMsg struct{}

// AppModel is the top-level Bubble Tea model.
type AppModel struct {
	// Layout
	Width  int
	Height int

	// Chrome components
	Info   InfoModel
	Menu   MenuModel
	Crumbs CrumbsModel
	Flash  FlashModel
	Filter FilterModel

	// Content
	ViewMode ViewMode
	Resource model.ResourceType
	Table    TableView
	Log      LogView
	Detail   DetailView

	// Navigation context (set on drill-down)
	SelectedProjectHash string
	SelectedSessionID   string
	SelectedAgentID     string

	// Data providers (injected from outside)
	DataProvider DataProvider

	// Animation tick counter
	tick int

	// Filter mode flag
	inFilter bool

	// State saved before a t/p/m jump (for esc-to-restore)
	jumpFrom *jumpFromState
}

// jumpFromState holds the navigation state before a t/p/m resource jump.
type jumpFromState struct {
	Resource            model.ResourceType
	SelectedProjectHash string
	SelectedSessionID   string
	SelectedAgentID     string
	Crumbs              CrumbsModel
	ViewMode            ViewMode
	NavItems            []MenuItem
	UtilItems           []MenuItem
}

// DataProvider is the interface for fetching resource data.
type DataProvider interface {
	GetProjects() any
	GetSessions(projectHash string) any
	GetAgents(sessionID string) any
	GetTools(agentID string) any
	GetTasks(sessionID string) any
	GetPlugins() any
	GetMCPServers() any
	// Navigation context
	CurrentProject() string
	CurrentSession() string
}

// NewAppModel creates a new application model.
func NewAppModel(dp DataProvider, initialResource model.ResourceType) AppModel {
	m := AppModel{
		DataProvider: dp,
		Resource:     initialResource,
	}
	m.Info = InfoModel{}
	m.Menu = MenuModel{
		NavItems:  TableNavItems(m.Resource),
		UtilItems: TableUtilItems(m.Resource),
	}
	m.Crumbs = CrumbsModel{Items: []string{string(initialResource)}}
	m.Flash = FlashModel{}
	m.Filter = FilterModel{}
	return m
}

// Init starts the tick timer.
func (m AppModel) Init() tea.Cmd {
	return tea.Batch(
		tick(),
		tea.EnterAltScreen,
	)
}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// Update handles messages.
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.updateSizes()
		return m, nil

	case TickMsg:
		m.tick++
		m.Flash.IsExpired() // lazy expiry check
		return m, tick()

	case tea.KeyMsg:
		// Ctrl+C always quits
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		// Filter mode
		if m.inFilter {
			return m.updateFilter(msg)
		}

		// Global keys (work in all view modes)
		switch msg.String() {
		case "/":
			m.inFilter = true
			m.Filter.Activate()
			return m, nil
		case "t":
			m.jumpTo(model.ResourceTasks)
			return m, nil
		case "p":
			m.jumpTo(model.ResourcePlugins)
			return m, nil
		case "m":
			m.jumpTo(model.ResourceMCP)
			return m, nil
		}

		// View-specific keys
		switch m.ViewMode {
		case ModeTable:
			return m.updateTable(msg)
		case ModeLog:
			return m.updateLog(msg)
		case ModeDetail, ModeYAML:
			return m.updateDetail(msg)
		}
	}

	return m, nil
}

func (m AppModel) updateFilter(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.inFilter = false
		m.Filter.Deactivate()
		m.Filter.Input = ""
		m.Table.Filter = ""
		m.Log.Filter = ""
	case "enter":
		m.inFilter = false
		m.Filter.Deactivate()
	case "backspace":
		m.Filter.Backspace()
		m.Table.Filter = m.Filter.Input
		m.Table.Selected = 0
		m.Log.Filter = m.Filter.Input
	default:
		if len(msg.Runes) == 1 {
			m.Filter.AddChar(msg.Runes[0])
			m.Table.Filter = m.Filter.Input
			m.Table.Selected = 0
			m.Log.Filter = m.Filter.Input
		}
	}
	return m, nil
}

func (m AppModel) updateTable(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		if m.Table.Filter != "" {
			m.Table.Filter = ""
			m.Filter.Input = ""
		} else {
			m.navigateBack()
		}
	case "enter":
		m.drillDown()
	case "l":
		if ResourceHasLog(m.Resource) {
			m.ViewMode = ModeLog
			m.Menu.NavItems = LogNavItems()
			m.Menu.UtilItems = LogUtilItems()
			m.refreshLog()
			return m, func() tea.Msg { return LogRequestMsg{} }
		}
	case "d":
		m.ViewMode = ModeDetail
		m.Menu.NavItems = DetailNavItems()
		m.Menu.UtilItems = DetailUtilItems()
		m.refreshDetail()
		return m, func() tea.Msg { return DetailRequestMsg{} }
	case "y":
		m.ViewMode = ModeYAML
		m.Menu.NavItems = DetailNavItems()
		m.Menu.UtilItems = DetailUtilItems()
		m.refreshDetail()
		return m, func() tea.Msg { return YAMLRequestMsg{} }
	default:
		m.Table.Update(msg)
	}
	return m, nil
}

// jumpTo switches to a flat resource, saving the current state for esc-restore.
func (m *AppModel) jumpTo(rt model.ResourceType) {
	m.jumpFrom = &jumpFromState{
		Resource:            m.Resource,
		SelectedProjectHash: m.SelectedProjectHash,
		SelectedSessionID:   m.SelectedSessionID,
		SelectedAgentID:     m.SelectedAgentID,
		Crumbs:              m.Crumbs,
		ViewMode:            m.ViewMode,
		NavItems:            m.Menu.NavItems,
		UtilItems:           m.Menu.UtilItems,
	}
	m.Resource = rt
	m.ViewMode = ModeTable
	m.Menu.NavItems = TableNavItems(rt)
	m.Menu.UtilItems = TableUtilItems(rt)
	m.Crumbs.Reset(string(rt))
	m.Filter.Deactivate()
	m.Filter.Input = ""
	m.Table.Filter = ""
}

func (m AppModel) updateLog(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" {
		m.ViewMode = ModeTable
		m.Menu.NavItems = TableNavItems(m.Resource)
		m.Menu.UtilItems = TableUtilItems(m.Resource)
		return m, nil
	}
	m.Log.Update(msg)
	return m, nil
}

func (m AppModel) updateDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" {
		m.ViewMode = ModeTable
		m.Menu.NavItems = TableNavItems(m.Resource)
		m.Menu.UtilItems = TableUtilItems(m.Resource)
		return m, nil
	}
	m.Detail.Update(msg)
	return m, nil
}

func (m *AppModel) navigateBack() {
	if m.ViewMode != ModeTable {
		m.ViewMode = ModeTable
		m.Menu.NavItems = TableNavItems(m.Resource)
		m.Menu.UtilItems = TableUtilItems(m.Resource)
		return
	}
	// Flat resources jumped to via t/p/m: restore previous state
	switch m.Resource {
	case model.ResourceTasks, model.ResourcePlugins, model.ResourceMCP:
		if m.jumpFrom != nil {
			m.Resource = m.jumpFrom.Resource
			m.SelectedProjectHash = m.jumpFrom.SelectedProjectHash
			m.SelectedSessionID = m.jumpFrom.SelectedSessionID
			m.SelectedAgentID = m.jumpFrom.SelectedAgentID
			m.Crumbs = m.jumpFrom.Crumbs
			m.ViewMode = m.jumpFrom.ViewMode
			m.Menu.NavItems = m.jumpFrom.NavItems
			m.Menu.UtilItems = m.jumpFrom.UtilItems
			m.jumpFrom = nil
		} else {
			// No saved state (e.g. started directly on this resource): go to projects.
			m.Resource = model.ResourceProjects
			m.SelectedProjectHash = ""
			m.SelectedSessionID = ""
			m.SelectedAgentID = ""
			m.Crumbs.Reset(string(model.ResourceProjects))
			m.ViewMode = ModeTable
			m.Menu.NavItems = TableNavItems(model.ResourceProjects)
			m.Menu.UtilItems = TableUtilItems(model.ResourceProjects)
		}
		return
	}
	// Navigate up the resource hierarchy
	switch m.Resource {
	case model.ResourceTools:
		m.SelectedAgentID = ""
		m.Crumbs.Pop()
		m.Resource = model.ResourceAgents
		m.ViewMode = ModeTable
		m.Menu.NavItems = TableNavItems(m.Resource)
		m.Menu.UtilItems = TableUtilItems(m.Resource)
	case model.ResourceAgents:
		m.SelectedSessionID = ""
		m.Crumbs.Pop()
		m.Resource = model.ResourceSessions
		m.ViewMode = ModeTable
		m.Menu.NavItems = TableNavItems(m.Resource)
		m.Menu.UtilItems = TableUtilItems(m.Resource)
	case model.ResourceSessions:
		m.SelectedProjectHash = ""
		m.Crumbs.Pop()
		m.Resource = model.ResourceProjects
		m.ViewMode = ModeTable
		m.Menu.NavItems = TableNavItems(m.Resource)
		m.Menu.UtilItems = TableUtilItems(m.Resource)
	}
}

func (m *AppModel) drillDown() {
	row := m.Table.SelectedRow()
	if row == nil {
		return
	}
	switch m.Resource {
	case model.ResourceProjects:
		if p, ok := row.Data.(*model.Project); ok {
			m.SelectedProjectHash = p.Hash
		}
		m.drillInto(model.ResourceSessions)
	case model.ResourceSessions:
		if s, ok := row.Data.(*model.Session); ok {
			m.SelectedSessionID = s.ID
		}
		m.drillInto(model.ResourceAgents)
	case model.ResourceAgents:
		if a, ok := row.Data.(*model.Agent); ok {
			m.SelectedAgentID = a.ID
		}
		m.drillInto(model.ResourceTools)
	}
}

func (m *AppModel) drillInto(rt model.ResourceType) {
	m.Resource = rt
	m.ViewMode = ModeTable
	m.Menu.NavItems = TableNavItems(m.Resource)
	m.Menu.UtilItems = TableUtilItems(m.Resource)
	m.Crumbs.Push(string(rt))
	m.Filter.Deactivate()
	m.Filter.Input = ""
}

func (m *AppModel) refreshLog() {
	m.Log = NewLogView(string(m.Resource), m.contentWidth(), m.contentHeight())
}

func (m *AppModel) refreshDetail() {
	m.Detail = NewDetailView(string(m.Resource), m.contentWidth(), m.contentHeight())
}

func (m *AppModel) updateSizes() {
	w := m.contentWidth()
	h := m.contentHeight()
	m.Info.Width = w
	m.Crumbs.Width = w
	m.Flash.Width = w
	m.Filter.Width = w
	m.Table.Width = w
	m.Table.Height = h
	m.Log.Width = w
	m.Log.Height = h
	m.Detail.Width = w
	m.Detail.Height = h
}

// ContentHeight returns the number of terminal lines available for content.
func (m AppModel) ContentHeight() int {
	return m.contentHeight()
}

func (m AppModel) contentHeight() int {
	// Sum chrome heights dynamically so adding/removing UI rows only
	// requires updating the relevant component's Height() method.
	chrome := m.Info.Height(len(m.Menu.NavItems), len(m.Menu.UtilItems)) + // info panel (top)
		1 + // title bar (renderTitleBar → always 1 line)
		m.Crumbs.Height() + // breadcrumb bar
		m.Flash.Height() // status bar (Flash and Filter render the same 1 line)
	h := m.Height - chrome
	if h < 5 {
		return 5
	}
	return h
}

func (m AppModel) contentWidth() int {
	if m.Width < 10 {
		return 10
	}
	return m.Width
}

// View renders the full UI in k9s-style: info panel + title bar + content + status.
func (m AppModel) View() string {
	if m.Width == 0 {
		return "Loading..."
	}

	// --- 1. Info panel (7 lines) ---
	infoStr := m.Info.ViewWithMenu(m.Menu.NavItems, m.Menu.UtilItems)

	// --- 2. Resource title bar ---
	titleStr := m.renderTitleBar()

	// --- 3. Content ---
	var contentStr string
	switch m.ViewMode {
	case ModeTable:
		contentStr = m.Table.View()
	case ModeLog:
		contentStr = m.Log.View()
	case ModeDetail, ModeYAML:
		contentStr = m.Detail.View()
	}
	rawLines := strings.Split(strings.TrimRight(contentStr, "\n"), "\n")
	if limit := m.contentHeight(); len(rawLines) > limit {
		rawLines = rawLines[:limit]
	}

	// --- 4. Status bar ---
	var statusView string
	if m.inFilter {
		statusView = m.Filter.View()
	} else {
		statusView = m.Flash.View()
	}

	// --- Assemble ---
	var sb strings.Builder
	sb.WriteString(infoStr + "\n")
	sb.WriteString(titleStr + "\n")
	for _, line := range rawLines {
		sb.WriteString(line + "\n")
	}
	sb.WriteString(m.Crumbs.View() + "\n")
	sb.WriteString(statusView)

	return sb.String()
}

// renderTitleBar renders the centered resource title line, e.g. "─── Sessions(all)[3] ───".
func (m AppModel) renderTitleBar() string {
	res := string(m.Resource)
	if len(res) > 0 {
		res = strings.ToUpper(res[:1]) + res[1:]
	}

	filter := "all"
	if m.Table.Filter != "" {
		filter = m.Table.Filter
	}

	count := m.Table.FilteredCount()
	title := fmt.Sprintf("%s(%s)[%d]", res, filter, count)
	titleStyled := StyleTitle.Render(title)
	titleVis := lipgloss.Width(titleStyled)

	gray := lipgloss.NewStyle().Foreground(colorGray)
	inner := max(
		// 2 for the spaces around title
		m.Width-titleVis-2, 0)
	leftDash := inner / 2
	rightDash := inner - leftDash

	return gray.Render(strings.Repeat("─", leftDash)+" ") + titleStyled + gray.Render(" "+strings.Repeat("─", rightDash))
}
