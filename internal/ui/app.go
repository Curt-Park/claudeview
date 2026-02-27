package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Curt-Park/claudeview/internal/model"
)

// TickMsg is sent on each timer tick for animations.
type TickMsg time.Time

// RefreshMsg signals data has been refreshed.
type RefreshMsg struct{}

// HighlightClearMsg clears the menu key highlight after HighlightDuration.
type HighlightClearMsg struct{}

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
	Resource model.ResourceType
	Table    TableView

	// Navigation context (set on drill-down)
	SelectedProjectHash string
	SelectedSessionID   string
	SelectedAgentID     string
	SelectedPlugin      *model.Plugin
	SelectedMemory      *model.Memory

	// Data providers (injected from outside)
	DataProvider DataProvider

	// Animation tick counter
	tick int

	// Filter mode flag
	inFilter bool

	// filterStack saves parent-view filters across drill-downs
	filterStack []string

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
	Filter              string
	FilterStack         []string
}

// DataProvider is the interface for fetching resource data.
type DataProvider interface {
	GetProjects() []*model.Project
	GetSessions(projectHash string) []*model.Session
	GetAgents(sessionID string) []*model.Agent
	GetPlugins(projectHash string) []*model.Plugin
	GetMemories(projectHash string) []*model.Memory
}

// NewAppModel creates a new application model.
func NewAppModel(dp DataProvider, initialResource model.ResourceType) AppModel {
	m := AppModel{
		DataProvider: dp,
		Resource:     initialResource,
	}
	m.Info = InfoModel{}
	m.refreshMenu()
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

	case HighlightClearMsg:
		m.Menu.ClearHighlight()
		return m, nil

	case tea.KeyMsg:
		// Ctrl+C always quits
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		// Highlight menu items only when not in filter input mode.
		var highlightCmd tea.Cmd
		if !m.inFilter {
			m.Menu.SetHighlight(msg.String())
			highlightCmd = tea.Tick(HighlightDuration, func(time.Time) tea.Msg {
				return HighlightClearMsg{}
			})
		}

		// Filter mode
		if m.inFilter {
			model, cmd := m.updateFilter(msg)
			return model, tea.Batch(cmd, highlightCmd)
		}

		// Global keys (work in all view modes)
		switch msg.String() {
		case "/":
			m.inFilter = true
			m.Filter.Activate()
			m.refreshMenu()
			return m, highlightCmd
		case "p":
			m.jumpTo(model.ResourcePlugins)
			return m, highlightCmd
		case "m":
			if m.SelectedProjectHash != "" {
				m.jumpTo(model.ResourceMemory)
			}
			return m, highlightCmd
		}

		// View-specific keys
		model, cmd := m.updateTable(msg)
		return model, tea.Batch(cmd, highlightCmd)
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
		m.refreshMenu()
	case "enter":
		m.inFilter = false
		m.Filter.Deactivate()
		m.refreshMenu()
	case "backspace":
		m.Filter.Backspace()
		m.Table.Filter = m.Filter.Input
		m.Table.Selected = 0
	default:
		if len(msg.Runes) == 1 {
			m.Filter.AddChar(msg.Runes[0])
			m.Table.Filter = m.Filter.Input
			m.Table.Selected = 0
		}
	}
	return m, nil
}

func (m AppModel) updateTable(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		if m.Table.Filter != "" {
			// Filter clear: same view stays — keep highlight, update menu desc.
			m.Table.Filter = ""
			m.Filter.Input = ""
			m.refreshMenu()
		} else {
			// View transition: clear highlight so next view doesn't show stale flash.
			m.Menu.ClearHighlight()
			m.navigateBack()
		}
	case "enter":
		// View transition: clear highlight so next view doesn't show stale flash.
		m.Menu.ClearHighlight()
		m.drillDown()
	default:
		m.Table.Update(msg)
	}
	return m, nil
}

// refreshMenu updates the menu nav and util items based on current state.
// hasFilter is true when a filter is active OR the filter input is open.
func (m *AppModel) refreshMenu() {
	m.Menu.NavItems = TableNavItems(m.Resource, m.Table.Filter != "" || m.inFilter)
	m.Menu.UtilItems = TableUtilItems(m.Resource)
}

// switchResource pops the breadcrumb, sets the resource, and refreshes the menu.
func (m *AppModel) switchResource(rt model.ResourceType) {
	m.Crumbs.Pop()
	m.Resource = rt
	m.refreshMenu()
}

// jumpTo switches to a flat resource, saving the current state for esc-restore.
func (m *AppModel) jumpTo(rt model.ResourceType) {
	m.jumpFrom = &jumpFromState{
		Resource:            m.Resource,
		SelectedProjectHash: m.SelectedProjectHash,
		SelectedSessionID:   m.SelectedSessionID,
		SelectedAgentID:     m.SelectedAgentID,
		Crumbs:              m.Crumbs,
		Filter:              m.Table.Filter,
		FilterStack:         m.filterStack,
	}
	m.Resource = rt
	m.filterStack = nil
	m.refreshMenu()
	m.Crumbs.Reset(string(rt))
	m.Filter.Deactivate()
	m.Filter.Input = ""
	m.Table.Filter = ""
}

// popFilter restores the parent view's filter from filterStack.
func (m *AppModel) popFilter() {
	if n := len(m.filterStack); n > 0 {
		m.Table.Filter = m.filterStack[n-1]
		m.Filter.Input = m.Table.Filter
		m.filterStack = m.filterStack[:n-1]
	} else {
		m.Table.Filter = ""
		m.Filter.Input = ""
	}
}

func (m *AppModel) navigateBack() {
	// Flat resources jumped to via t/p/m: restore previous state
	switch m.Resource {
	case model.ResourcePlugins, model.ResourceMemory:
		if m.jumpFrom != nil {
			m.Resource = m.jumpFrom.Resource
			m.SelectedProjectHash = m.jumpFrom.SelectedProjectHash
			m.SelectedSessionID = m.jumpFrom.SelectedSessionID
			m.SelectedAgentID = m.jumpFrom.SelectedAgentID
			m.Crumbs = m.jumpFrom.Crumbs
			m.Table.Filter = m.jumpFrom.Filter
			m.Filter.Input = m.jumpFrom.Filter
			m.filterStack = m.jumpFrom.FilterStack
			m.jumpFrom = nil
			m.refreshMenu()
		} else {
			// No saved state (e.g. started directly on this resource): go to projects.
			m.Resource = model.ResourceProjects
			m.SelectedProjectHash = ""
			m.SelectedSessionID = ""
			m.SelectedAgentID = ""
			m.Crumbs.Reset(string(model.ResourceProjects))
			m.refreshMenu()
		}
		return
	}
	// Navigate up the resource hierarchy
	switch m.Resource {
	case model.ResourceAgents:
		m.SelectedSessionID = ""
		m.popFilter()
		m.switchResource(model.ResourceSessions)
	case model.ResourceSessions:
		m.SelectedProjectHash = ""
		m.popFilter()
		m.switchResource(model.ResourceProjects)
	case model.ResourcePluginDetail:
		m.popFilter()
		m.switchResource(model.ResourcePlugins)
	case model.ResourceMemoryDetail:
		m.popFilter()
		m.switchResource(model.ResourceMemory)
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
	case model.ResourcePlugins:
		if p, ok := row.Data.(*model.Plugin); ok {
			m.SelectedPlugin = p
		}
		m.drillInto(model.ResourcePluginDetail)
	case model.ResourceMemory:
		if mem, ok := row.Data.(*model.Memory); ok {
			m.SelectedMemory = mem
		}
		m.drillInto(model.ResourceMemoryDetail)
	}
}

func (m *AppModel) drillInto(rt model.ResourceType) {
	m.filterStack = append(m.filterStack, m.Table.Filter)
	m.Table.Filter = ""
	m.Resource = rt
	m.Crumbs.Push(string(rt))
	m.Filter.Deactivate()
	m.Filter.Input = ""
	m.refreshMenu()
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
	infoStr := m.Info.ViewWithMenu(m.Menu)

	// --- 2. Resource title bar ---
	titleStr := m.renderTitleBar()

	// --- 3. Content ---
	var contentStr string
	switch m.Resource {
	case model.ResourcePluginDetail:
		contentStr = RenderPluginDetail(m.SelectedPlugin)
	case model.ResourceMemoryDetail:
		contentStr = RenderMemoryDetail(m.SelectedMemory)
	default:
		contentStr = m.Table.View()
	}
	rawLines := strings.Split(strings.TrimRight(contentStr, "\n"), "\n")
	limit := m.contentHeight()
	if len(rawLines) > limit {
		rawLines = rawLines[:limit]
	}
	for len(rawLines) < limit {
		rawLines = append(rawLines, "")
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
