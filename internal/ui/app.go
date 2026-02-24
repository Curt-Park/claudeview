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
)

// TickMsg is sent on each timer tick for animations.
type TickMsg time.Time

// RefreshMsg signals data has been refreshed.
type RefreshMsg struct{}

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

	// Navigation state
	Keys KeyMap

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
		Keys:         DefaultKeyMap(),
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
		case ModeDetail:
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
	case "esc", "enter":
		m.inFilter = false
		m.Filter.Deactivate()
	case "backspace":
		m.Filter.Backspace()
	default:
		if len(msg.Runes) == 1 {
			m.Filter.AddChar(msg.Runes[0])
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
	case "d":
		m.ViewMode = ModeDetail
		m.Menu.Items = DetailMenuItems()
		m.refreshDetail()
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
		m.switchResource(model.ResourceAgents)
	case model.ResourceAgents:
		m.switchResource(model.ResourceSessions)
	case model.ResourceSessions:
		m.switchResource(model.ResourceProjects)
	}
}

func (m *AppModel) drillDown() {
	switch m.Resource {
	case model.ResourceProjects:
		m.switchResource(model.ResourceSessions)
	case model.ResourceSessions:
		m.switchResource(model.ResourceAgents)
	case model.ResourceAgents:
		m.switchResource(model.ResourceTools)
	}
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
	m.Header.Width = m.Width
	m.Menu.Width = m.Width
	m.Crumbs.Width = m.Width
	m.Flash.Width = m.Width
	m.Command.Width = m.Width
	m.Filter.Width = m.Width
	m.Table.Width = m.Width
	m.Table.Height = m.contentHeight()
	m.Log.Width = m.contentWidth()
	m.Log.Height = m.contentHeight()
	m.Detail.Width = m.contentWidth()
	m.Detail.Height = m.contentHeight()
}

func (m AppModel) contentHeight() int {
	// header(1) + menu(1) + crumbs(1) + flash(1) = 4 chrome rows
	h := m.Height - 4
	if h < 5 {
		return 5
	}
	return h
}

func (m AppModel) contentWidth() int {
	return m.Width
}

// View renders the full UI.
func (m AppModel) View() string {
	if m.Width == 0 {
		return "Loading..."
	}

	var sections []string

	sections = append(sections, m.Header.View())
	sections = append(sections, m.Menu.View())

	// Content area
	var content string
	switch m.ViewMode {
	case ModeTable:
		content = m.Table.View()
	case ModeLog:
		content = m.Log.View()
	case ModeDetail:
		content = m.Detail.View()
	}
	sections = append(sections, content)

	sections = append(sections, m.Crumbs.View())

	// Bottom bar: flash or command or filter
	if m.inCommand {
		sections = append(sections, m.Command.View())
	} else if m.inFilter {
		sections = append(sections, m.Filter.View())
	} else {
		sections = append(sections, m.Flash.View())
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		strings.Join(sections, "\n"),
	)
}
