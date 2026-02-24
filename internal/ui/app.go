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
	Header  HeaderModel
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

	// Navigation context (set on drill-down)
	SelectedProjectHash string
	SelectedSessionID   string
	SelectedAgentID     string

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
	CurrentAgent() string
}

// NewAppModel creates a new application model.
func NewAppModel(dp DataProvider, initialResource model.ResourceType) AppModel {
	m := AppModel{
		DataProvider: dp,
		Resource:     initialResource,
	}
	m.Header = HeaderModel{}
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

	case RefreshMsg:
		m.refreshCurrentView()
		return m, nil

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
			m.showHelp()
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

func (m *AppModel) switchResource(rt model.ResourceType) {
	m.Resource = rt
	m.ViewMode = ModeTable
	m.Menu.Items = TableMenuItems()
	m.Crumbs.Reset(string(rt))
	m.Filter.Deactivate()
	m.Filter.Input = ""
	m.Flash.Set(fmt.Sprintf("switched to %s", rt), FlashInfo, 2*time.Second)
	m.refreshCurrentView()
}

func (m *AppModel) navigateBack() {
	if m.ViewMode != ModeTable {
		m.ViewMode = ModeTable
		m.Menu.Items = TableMenuItems()
		return
	}
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
	// Log view is populated by view layer
	m.Log = NewLogView(string(m.Resource), m.contentWidth(), m.contentHeight())
}

func (m *AppModel) refreshDetail() {
	m.Detail = NewDetailView(string(m.Resource), m.contentWidth(), m.contentHeight())
}

func (m *AppModel) refreshCurrentView() {
	// Trigger data refresh — actual implementation in cmd layer
}

func (m *AppModel) showHelp() {
	m.Flash.Set("?: help  q: quit  j/k: nav  enter: drill  l: logs  d: detail  :: command  /: filter", FlashInfo, 10*time.Second)
}

func (m *AppModel) updateSizes() {
	inner := m.contentWidth() // Width - 4 for bordered layout
	m.Header.Width = inner
	m.Menu.Width = inner
	m.Crumbs.Width = inner
	m.Flash.Width = inner
	m.Command.Width = inner
	m.Filter.Width = inner
	m.Table.Width = inner
	m.Table.Height = m.contentHeight()
	m.Log.Width = inner
	m.Log.Height = m.contentHeight()
	m.Detail.Width = inner
	m.Detail.Height = m.contentHeight()
}

func (m AppModel) contentHeight() int {
	// top-border(1) + menu(1) + sep(1) + sep(1) + crumbs(1) + status(1) + bottom(1) = 7 chrome rows
	h := m.Height - 7
	if h < 5 {
		return 5
	}
	return h
}

func (m AppModel) contentWidth() int {
	// subtract 4 for "│ " on the left and " │" on the right
	w := m.Width - 4
	if w < 10 {
		return 10
	}
	return w
}

// View renders the full UI with a bordered box layout.
func (m AppModel) View() string {
	if m.Width == 0 {
		return "Loading..."
	}

	bc := lipgloss.NewStyle().Foreground(colorGray)
	titleStyle := lipgloss.NewStyle().Foreground(colorWhite).Bold(true)
	innerW := m.contentWidth()

	// wrapLine adds "│ " prefix and " │" suffix around a pre-rendered inner line.
	wrapLine := func(s string) string {
		// Ensure inner content is exactly innerW visible chars.
		w := lipgloss.Width(s)
		if w < innerW {
			s += strings.Repeat(" ", innerW-w)
		}
		return bc.Render("│") + " " + s + " " + bc.Render("│")
	}

	hbar := func(left, right string) string {
		return bc.Render(left + strings.Repeat("─", m.Width-2) + right)
	}

	// --- Top border with header text integrated ---
	headerText := m.Header.ContentText()
	prefix := "┌─ "
	suffix := " ─┐"
	fillLen := max(0, m.Width-len([]rune(prefix))-lipgloss.Width(headerText)-len([]rune(suffix)))
	topBorder := bc.Render(prefix) + titleStyle.Render(headerText) +
		bc.Render(" "+strings.Repeat("─", fillLen)+"─┐")

	// --- Menu row ---
	menuLine := wrapLine(m.Menu.View())

	// --- Separator ---
	sep := hbar("├", "┤")

	// --- Content lines ---
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
	// Clip to the exact content height so the total view fits the terminal.
	// The table View() produces contentHeight+1 lines (col header + data rows);
	// clamping here prevents the top border from being scrolled off screen.
	if limit := m.contentHeight(); len(rawLines) > limit {
		rawLines = rawLines[:limit]
	}
	wrappedContent := make([]string, len(rawLines))
	for i, line := range rawLines {
		wrappedContent[i] = wrapLine(line)
	}

	// --- Bottom separator + crumbs + status ---
	botSep := hbar("├", "┤")
	crumbsLine := wrapLine(m.Crumbs.View())

	var statusView string
	if m.inCommand {
		statusView = m.Command.View()
	} else if m.inFilter {
		statusView = m.Filter.View()
	} else {
		statusView = m.Flash.View()
	}
	statusLine := wrapLine(statusView)

	botBorder := hbar("└", "┘")

	// --- Assemble ---
	var sb strings.Builder
	sb.WriteString(topBorder + "\n")
	sb.WriteString(menuLine + "\n")
	sb.WriteString(sep + "\n")
	for _, cl := range wrappedContent {
		sb.WriteString(cl + "\n")
	}
	sb.WriteString(botSep + "\n")
	sb.WriteString(crumbsLine + "\n")
	sb.WriteString(statusLine + "\n")
	sb.WriteString(botBorder)

	return sb.String()
}
