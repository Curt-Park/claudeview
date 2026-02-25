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
	ModeHelp            // ? key — full-screen help overlay
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
	Info    InfoModel
	Menu    MenuModel
	Crumbs  CrumbsModel
	Flash   FlashModel
	Command CommandModel
	Filter  FilterModel

	// Content
	ViewMode ViewMode
	Resource model.ResourceType
	Table    TableView
	Log      LogView
	Detail   DetailView
	Help     HelpView

	// Navigation context (set on drill-down)
	SelectedProjectHash string
	SelectedSessionID   string
	SelectedAgentID     string

	// ParentFilter is the label of the currently selected parent shortcut (0=all means "").
	ParentFilter string

	// Data providers (injected from outside)
	DataProvider DataProvider

	// Animation tick counter
	tick int

	// Command/filter mode flags
	inCommand bool
	inFilter  bool
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
	m.Menu = MenuModel{Items: TableMenuItems()}
	m.Crumbs = CrumbsModel{Items: []string{string(initialResource)}}
	m.Flash = FlashModel{}
	m.Command = CommandModel{}
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

		// Command mode
		if m.inCommand {
			return m.updateCommand(msg)
		}

		// Filter mode
		if m.inFilter {
			return m.updateFilter(msg)
		}

		// Global keys
		switch msg.String() {
		case ":":
			m.inCommand = true
			m.Command.Activate()
			return m, nil
		case "/":
			m.inFilter = true
			m.Filter.Activate()
			return m, nil
		case "?":
			m.ViewMode = ModeHelp
			m.Help = NewHelpView(m.contentWidth(), m.contentHeight())
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
		case ModeHelp:
			return m.updateHelp(msg)
		}
	}

	return m, nil
}

func (m AppModel) updateCommand(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.inCommand = false
		m.Command.Deactivate()
	case "enter":
		rt, ok := m.Command.Submit()
		m.inCommand = false
		m.Command.Deactivate()
		if ok {
			m.switchResource(rt)
		} else {
			m.Flash.Set(fmt.Sprintf("unknown resource: %s", m.Command.Input), FlashError, 3*time.Second)
		}
	case "tab":
		m.Command.Input = m.Command.Accept()
	case "backspace":
		m.Command.Backspace()
	default:
		if len(msg.Runes) == 1 {
			m.Command.AddChar(msg.Runes[0])
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
		m.Log.Filter = m.Filter.Input
	default:
		if len(msg.Runes) == 1 {
			m.Filter.AddChar(msg.Runes[0])
			m.Table.Filter = m.Filter.Input
			m.Log.Filter = m.Filter.Input
		}
	}
	return m, nil
}

func (m AppModel) updateTable(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.navigateBack()
	case "enter":
		m.drillDown()
	case "l":
		m.ViewMode = ModeLog
		m.Menu.Items = LogMenuItems()
		m.refreshLog()
		return m, func() tea.Msg { return LogRequestMsg{} }
	case "d":
		m.ViewMode = ModeDetail
		m.Menu.Items = DetailMenuItems()
		m.refreshDetail()
		return m, func() tea.Msg { return DetailRequestMsg{} }
	case "y":
		m.ViewMode = ModeYAML
		m.Menu.Items = DetailMenuItems()
		m.refreshDetail()
		return m, func() tea.Msg { return YAMLRequestMsg{} }
	case "0":
		// Clear parent filter — show all
		m.Table.Filter = ""
		m.Filter.Input = ""
		m.ParentFilter = ""
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		idx := int(msg.Runes[0] - '1')
		if idx < len(m.Info.ParentShortcuts) {
			sc := m.Info.ParentShortcuts[idx]
			m.ParentFilter = sc.Label
			// Apply as table filter so rows are filtered by parent label
			m.Table.Filter = sc.Label
			m.Filter.Input = sc.Label
		}
	default:
		m.Table.Update(msg)
	}
	return m, nil
}

func (m AppModel) updateLog(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" {
		m.ViewMode = ModeTable
		m.Menu.Items = TableMenuItems()
		return m, nil
	}
	m.Log.Update(msg)
	return m, nil
}

func (m AppModel) updateDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" {
		m.ViewMode = ModeTable
		m.Menu.Items = TableMenuItems()
		return m, nil
	}
	m.Detail.Update(msg)
	return m, nil
}

func (m AppModel) updateHelp(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" || msg.String() == "q" || msg.String() == "?" {
		m.ViewMode = ModeTable
		m.Menu.Items = TableMenuItems()
		return m, nil
	}
	m.Help.Update(msg)
	return m, nil
}

func (m *AppModel) switchResource(rt model.ResourceType) {
	m.Resource = rt
	// Reset navigation context for flat (unfiltered) access via :command
	m.SelectedProjectHash = ""
	m.SelectedSessionID = ""
	m.SelectedAgentID = ""
	m.ParentFilter = ""
	m.ViewMode = ModeTable
	m.Menu.Items = TableMenuItems()
	m.Crumbs.Reset(string(rt))
	m.Filter.Deactivate()
	m.Filter.Input = ""
	m.Flash.Set(fmt.Sprintf("switched to %s", rt), FlashInfo, 2*time.Second)

}

func (m *AppModel) navigateBack() {
	if m.ViewMode != ModeTable {
		m.ViewMode = ModeTable
		m.Menu.Items = TableMenuItems()
		return
	}
	m.ParentFilter = ""
	// Navigate up the resource hierarchy
	switch m.Resource {
	case model.ResourceTools:
		m.SelectedAgentID = ""
		m.Crumbs.Pop()
		m.Resource = model.ResourceAgents
		m.ViewMode = ModeTable
		m.Menu.Items = TableMenuItems()
	case model.ResourceAgents:
		m.SelectedSessionID = ""
		m.Crumbs.Pop()
		m.Resource = model.ResourceSessions
		m.ViewMode = ModeTable
		m.Menu.Items = TableMenuItems()
	case model.ResourceSessions:
		m.SelectedProjectHash = ""
		m.Crumbs.Pop()
		m.Resource = model.ResourceProjects
		m.ViewMode = ModeTable
		m.Menu.Items = TableMenuItems()
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
	m.Menu.Items = TableMenuItems()
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
	m.Command.Width = w
	m.Filter.Width = w
	m.Table.Width = w
	m.Table.Height = h
	m.Log.Width = w
	m.Log.Height = h
	m.Detail.Width = w
	m.Detail.Height = h
	m.Help.Width = w
	m.Help.Height = h
}

func (m AppModel) contentHeight() int {
	// 7 info rows + 1 title bar + 1 crumbs + 1 status = 10 chrome rows
	h := m.Height - 10
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
	infoStr := m.Info.ViewWithMenu(m.Menu.Items)

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
	case ModeHelp:
		contentStr = m.Help.View()
	}
	rawLines := strings.Split(strings.TrimRight(contentStr, "\n"), "\n")
	if limit := m.contentHeight(); len(rawLines) > limit {
		rawLines = rawLines[:limit]
	}

	// --- 4. Status bar ---
	var statusView string
	if m.inCommand {
		statusView = m.Command.View()
	} else if m.inFilter {
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
