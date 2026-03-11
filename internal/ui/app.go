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

// SyncViewMsg signals that the chat table needs to be re-synced immediately
// (e.g. after expand/collapse of a ChatItem's tool call sub-rows).
type SyncViewMsg struct{}

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
	Resource      model.ResourceType
	Table         TableView
	ContentOffset int // scroll offset for content-only views

	// Navigation context (set on drill-down)
	SelectedProjectHash string
	SelectedSessionID   string
	SelectedPlugin      *model.Plugin
	SelectedPluginItem  *model.PluginItem
	SelectedMemory      *model.Memory

	// Session chat data (set on drill-down into session-chat)
	SelectedTurns              []model.Turn
	SubagentTurns              [][]model.Turn
	SubagentTypes              []model.AgentType
	ChatFollow                 bool   // true = auto-scroll to bottom (tail -f mode)
	SelectedSessionSlug        string // slug for the selected session (shown in header)
	SelectedSessionFilePath    string // for async reload
	SelectedSessionSubagentDir string // for async subagent reload

	// Slug group: all sessions in the selected slug group (len > 1 when merged)
	SlugSessions      []*model.Session
	slugGroupTurns    [][]model.Turn      // per-session turns for slug group
	slugGroupSubTurns [][][]model.Turn    // per-session subagent turns
	slugGroupSubTypes [][]model.AgentType // per-session subagent types

	// Chat table state
	ChatItems        []ChatItem      // flattened selectable items
	SelectedChatItem int             // index of expanded item in detail view
	ExpandedItems    map[string]bool // ChatItemKey → expanded (tool call sub-rows visible)
	SelectedToolCall *ToolCallRow    // for tool-call-detail view

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
	Crumbs              CrumbsModel
	Filter              string
	FilterStack         []string
}

// isSubView returns true for views nested under plugins or memory
// (plugin-detail, plugin-item-detail, memory-detail).
// p/m jump keys are blocked in these views to preserve navigation context.
func isSubView(rt model.ResourceType) bool {
	return rt == model.ResourcePluginDetail ||
		rt == model.ResourcePluginItemDetail ||
		rt == model.ResourceMemoryDetail ||
		rt == model.ResourceHistoryDetail ||
		rt == model.ResourceToolCallDetail
}

// isContentView returns true for views that render flat text (not a table).
// These views use ContentOffset for scrolling instead of Table navigation.
func isContentView(rt model.ResourceType) bool {
	return rt == model.ResourcePluginItemDetail ||
		rt == model.ResourceMemoryDetail ||
		rt == model.ResourceHistoryDetail ||
		rt == model.ResourceToolCallDetail
}

// DataProvider is the interface for fetching resource data.
type DataProvider interface {
	GetProjects() []*model.Project
	GetSessions(projectHash string) []*model.Session
	GetAgents(sessionID string) []*model.Agent
	GetPlugins(projectHash string) []*model.Plugin
	GetPluginItems(plugin *model.Plugin) []*model.PluginItem
	GetMemories(projectHash string) []*model.Memory
	GetTurns(filePath string) []model.Turn
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
			highlightKey := msg.String()
			if highlightKey == " " {
				highlightKey = "space"
			}
			m.Menu.SetHighlight(highlightKey)
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
			if m.Resource != model.ResourceMemoryDetail && m.Resource != model.ResourcePluginItemDetail &&
				m.Resource != model.ResourceHistoryDetail && m.Resource != model.ResourceToolCallDetail {
				m.inFilter = true
				m.Filter.Activate()
				m.refreshMenu()
			}
			return m, highlightCmd
		case "p":
			if !isSubView(m.Resource) {
				m.jumpTo(model.ResourcePlugins)
			}
			return m, highlightCmd
		case "m":
			if m.SelectedProjectHash != "" && !isSubView(m.Resource) {
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
		if m.Resource == model.ResourceHistory {
			m.drillDetailFromRow()
		} else {
			return m, m.drillDown()
		}
	case " ":
		if m.Resource == model.ResourceHistory {
			return m, m.toggleExpansion()
		}
	default:
		if isContentView(m.Resource) {
			m.updateContentScroll(msg)
		} else {
			m.updateChatFollow(msg)
			m.Table.Update(msg)
			m.refreshMenu()
		}
	}
	return m, nil
}

// contentMaxOffset returns the maximum scroll offset for the current content view.
func (m AppModel) contentMaxOffset() int {
	var contentStr string
	switch m.Resource {
	case model.ResourcePluginItemDetail:
		contentStr = RenderPluginItemDetail(m.SelectedPluginItem, m.contentWidth())
	case model.ResourceMemoryDetail:
		contentStr = RenderMemoryDetail(m.SelectedMemory, m.contentWidth())
	case model.ResourceHistoryDetail:
		contentStr = RenderChatItemDetail(m.ChatItems, m.SelectedChatItem, m.contentWidth())
	case model.ResourceToolCallDetail:
		contentStr = RenderToolCallDetail(m.SelectedToolCall, m.contentWidth())
	default:
		return 0
	}
	lines := strings.Split(strings.TrimRight(contentStr, "\n"), "\n")
	if max := len(lines) - m.contentHeight(); max > 0 {
		return max
	}
	return 0
}

// updateContentScroll handles movement keys for content-only views (plugin-item-detail,
// memory-detail, session-chat-detail). It adjusts ContentOffset, capped to the actual
// content length.
func (m *AppModel) updateContentScroll(msg tea.KeyMsg) {
	half := m.contentHeight() / 2
	cap := func() {
		if max := m.contentMaxOffset(); m.ContentOffset > max {
			m.ContentOffset = max
		}
	}
	switch msg.String() {
	case "j", "down":
		m.ContentOffset++
		cap()
	case "k", "up":
		if m.ContentOffset > 0 {
			m.ContentOffset--
		}
	case "G":
		m.ContentOffset = m.contentMaxOffset()
	case "g":
		m.ContentOffset = 0
	case "ctrl+d", "pgdown":
		m.ContentOffset += half
		cap()
	case "ctrl+u", "pgup":
		if m.ContentOffset >= half {
			m.ContentOffset -= half
		} else {
			m.ContentOffset = 0
		}
	}
}

// updateChatFollow manages ChatFollow state based on navigation keys.
// G enables follow; k/g/ctrl+u disable it.
func (m *AppModel) updateChatFollow(msg tea.KeyMsg) {
	if m.Resource != model.ResourceHistory {
		return
	}
	switch msg.String() {
	case "G":
		m.ChatFollow = true
	case "k", "up", "g", "ctrl+u", "pgup":
		m.ChatFollow = false
	case "j", "down", "ctrl+d", "pgdown":
		// Enable follow if we land on the last row
		n := m.Table.FilteredCount()
		if n > 0 && m.Table.Selected >= n-2 {
			m.ChatFollow = true
		}
	}
}

// RebuildChatItems rebuilds the flattened ChatItems list from current turn data.
// In history-detail, re-resolves SelectedChatItem and adjusts ContentOffset
// so the view doesn't jump when ExtraTurns grouping changes on refresh.
func (m *AppModel) RebuildChatItems() {
	// Save the selected item's identity before rebuild.
	var prevKey string
	inDetail := m.Resource == model.ResourceHistoryDetail
	if inDetail && m.SelectedChatItem >= 0 && m.SelectedChatItem < len(m.ChatItems) {
		prevKey = ChatItemKey(m.ChatItems[m.SelectedChatItem])
	}

	if len(m.SlugSessions) > 1 {
		ids := make([]string, len(m.SlugSessions))
		for i, s := range m.SlugSessions {
			ids[i] = s.ShortID()
		}
		m.ChatItems = BuildMergedChatItems(m.slugGroupTurns, m.slugGroupSubTurns, m.slugGroupSubTypes, ids)
	} else {
		m.ChatItems = BuildChatItems(m.SelectedTurns, m.SubagentTurns, m.SubagentTypes)
	}

	// Re-resolve the selected item so it stays on the same turn after regrouping.
	if inDetail && prevKey != "" {
		for i, item := range m.ChatItems {
			if ChatItemKey(item) == prevKey {
				m.SelectedChatItem = i
				break
			}
		}
	}
}

// loadSlugGroupTurns loads turns from all sessions in the slug group.
func (m *AppModel) loadSlugGroupTurns() {
	m.slugGroupTurns = nil
	m.slugGroupSubTurns = nil
	m.slugGroupSubTypes = nil
	for _, s := range m.SlugSessions {
		turns := m.DataProvider.GetTurns(s.FilePath)
		m.slugGroupTurns = append(m.slugGroupTurns, turns)
		agents := m.DataProvider.GetAgents(s.ID)
		var subTurns [][]model.Turn
		var subTypes []model.AgentType
		for _, a := range agents {
			if a.IsSubagent && a.FilePath != "" {
				subTurns = append(subTurns, m.DataProvider.GetTurns(a.FilePath))
				subTypes = append(subTypes, a.Type)
			}
		}
		m.slugGroupSubTurns = append(m.slugGroupSubTurns, subTurns)
		m.slugGroupSubTypes = append(m.slugGroupSubTypes, subTypes)
	}
}

// SetSlugGroupData updates the slug group turn data (used by async reload).
func (m *AppModel) SetSlugGroupData(turns [][]model.Turn, subTurns [][][]model.Turn, subTypes [][]model.AgentType) {
	m.slugGroupTurns = turns
	m.slugGroupSubTurns = subTurns
	m.slugGroupSubTypes = subTypes
}

// refreshMenu updates the menu nav and util items based on current state.
// hasFilter is true when a filter is active OR the filter input is open.
// RefreshMenu is the exported entry point for callers outside this package (e.g. cmd/root.go)
// that need to sync the menu after mutating Table state (expansion, cursor restoration).
func (m *AppModel) RefreshMenu() { m.refreshMenu() }

func (m *AppModel) refreshMenu() {
	hasFilter := m.Table.Filter != "" || m.inFilter
	canExpand := false
	if row := m.Table.SelectedRow(); row != nil {
		if ci, ok := row.Data.(ChatItem); ok && !ci.IsDivider && len(ci.AllToolCalls()) > 0 {
			canExpand = true
		}
	}
	m.Menu.NavItems = TableNavItems(m.Resource, hasFilter)
	m.Menu.ActionItems = TableActionItems(m.Resource, hasFilter, canExpand)
	m.Menu.UtilItems = TableUtilItems(m.Resource, hasFilter)
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
	m.ContentOffset = 0
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
			m.Crumbs = m.jumpFrom.Crumbs
			m.Table.Filter = m.jumpFrom.Filter
			m.Filter.Input = m.jumpFrom.Filter
			m.filterStack = m.jumpFrom.FilterStack
			m.jumpFrom = nil
			m.ContentOffset = 0
			m.refreshMenu()
		} else {
			// No saved state (e.g. started directly on this resource): go to projects.
			m.Resource = model.ResourceProjects
			m.SelectedProjectHash = ""
			m.SelectedSessionID = ""
			m.Crumbs.Reset(string(model.ResourceProjects))
			m.ContentOffset = 0
			m.refreshMenu()
		}
		return
	}
	// Navigate up the resource hierarchy
	m.ContentOffset = 0
	switch m.Resource {
	case model.ResourceToolCallDetail:
		m.switchResource(model.ResourceHistory)
	case model.ResourceHistoryDetail:
		m.switchResource(model.ResourceHistory)
	case model.ResourceHistory:
		m.ExpandedItems = nil
		m.SelectedSessionID = ""
		m.SelectedSessionSlug = ""
		m.SelectedSessionFilePath = ""
		m.SelectedSessionSubagentDir = ""
		m.SelectedTurns = nil
		m.SubagentTurns = nil
		m.SubagentTypes = nil
		m.SlugSessions = nil
		m.slugGroupTurns = nil
		m.slugGroupSubTurns = nil
		m.slugGroupSubTypes = nil
		m.ChatFollow = false
		m.ChatItems = nil
		m.popFilter()
		m.switchResource(model.ResourceSessions)
	case model.ResourceSessions:
		m.SelectedProjectHash = ""
		m.popFilter()
		m.switchResource(model.ResourceProjects)
	case model.ResourcePluginDetail:
		m.popFilter()
		m.switchResource(model.ResourcePlugins)
	case model.ResourcePluginItemDetail:
		m.popFilter()
		m.switchResource(model.ResourcePluginDetail)
	case model.ResourceMemoryDetail:
		m.popFilter()
		m.switchResource(model.ResourceMemory)
	}
}

// toggleExpansion expands or collapses the selected ChatItem's tool call sub-rows.
// No-ops for dividers, user/assistant rows without tool calls, and ToolCallRow sub-rows.
func (m *AppModel) toggleExpansion() tea.Cmd {
	row := m.Table.SelectedRow()
	if row == nil {
		return nil
	}
	ci, ok := row.Data.(ChatItem)
	if !ok || ci.IsDivider || len(ci.AllToolCalls()) == 0 {
		return nil
	}
	key := ChatItemKey(ci)
	if m.ExpandedItems == nil {
		m.ExpandedItems = make(map[string]bool)
	}
	if m.ExpandedItems[key] {
		delete(m.ExpandedItems, key)
	} else {
		m.ExpandedItems[key] = true
	}
	m.ChatFollow = false
	return func() tea.Msg { return SyncViewMsg{} }
}

func (m *AppModel) drillDown() tea.Cmd {
	row := m.Table.SelectedRow()
	if row == nil {
		return nil
	}
	switch m.Resource {
	case model.ResourceHistory:
		// Tool call sub-row → tool call detail view
		if tr, ok := row.Data.(ToolCallRow); ok {
			m.SelectedToolCall = &tr
			m.drillInto(model.ResourceToolCallDetail)
			return nil
		}
		if ci, ok := row.Data.(ChatItem); ok {
			if ci.IsDivider {
				return nil
			}
			// No tool calls → full detail directly
			if len(ci.AllToolCalls()) == 0 {
				m.drillDetail(ci)
				return nil
			}
			// Has tool calls → expand/collapse
			key := ChatItemKey(ci)
			if m.ExpandedItems == nil {
				m.ExpandedItems = make(map[string]bool)
			}
			if m.ExpandedItems[key] {
				delete(m.ExpandedItems, key)
			} else {
				m.ExpandedItems[key] = true
			}
			m.ChatFollow = false
			return func() tea.Msg { return SyncViewMsg{} }
		}
		return nil
	case model.ResourceProjects:
		if p, ok := row.Data.(*model.Project); ok {
			m.SelectedProjectHash = p.Hash
		}
		m.drillInto(model.ResourceSessions)
	case model.ResourceSessions:
		if s, ok := row.Data.(*model.Session); ok {
			m.SelectedSessionID = s.ID
			m.SelectedSessionSlug = s.Slug
			m.SelectedSessionFilePath = s.FilePath
			m.SelectedSessionSubagentDir = s.SubagentDir

			if s.IsGroupRepresentative() {
				m.SlugSessions = s.GroupSessions
				m.loadSlugGroupTurns()
			} else {
				m.SlugSessions = nil
				m.SelectedTurns = m.DataProvider.GetTurns(s.FilePath)
				agents := m.DataProvider.GetAgents(s.ID)
				m.SubagentTurns = nil
				m.SubagentTypes = nil
				for _, a := range agents {
					if a.IsSubagent && a.FilePath != "" {
						m.SubagentTurns = append(m.SubagentTurns, m.DataProvider.GetTurns(a.FilePath))
						m.SubagentTypes = append(m.SubagentTypes, a.Type)
					}
				}
			}
		}
		m.drillInto(model.ResourceHistory)
		m.RebuildChatItems()
	case model.ResourcePlugins:
		if p, ok := row.Data.(*model.Plugin); ok {
			m.SelectedPlugin = p
		}
		m.drillInto(model.ResourcePluginDetail)
	case model.ResourcePluginDetail:
		if pi, ok := row.Data.(*model.PluginItem); ok {
			m.SelectedPluginItem = pi
		}
		m.drillInto(model.ResourcePluginItemDetail)
	case model.ResourceMemory:
		if mem, ok := row.Data.(*model.Memory); ok {
			m.SelectedMemory = mem
		}
		m.drillInto(model.ResourceMemoryDetail)
	}
	return nil
}

// drillDetail navigates to the full history detail view for the given ChatItem.
func (m *AppModel) drillDetail(ci ChatItem) {
	ciKey := ChatItemKey(ci)
	for i, item := range m.ChatItems {
		if ChatItemKey(item) == ciKey {
			m.SelectedChatItem = i
			break
		}
	}
	m.drillInto(model.ResourceHistoryDetail)
}

// drillDetailFromRow navigates to the appropriate detail view for the currently
// selected history row: message detail for ChatItem rows, tool call detail for sub-rows.
func (m *AppModel) drillDetailFromRow() {
	row := m.Table.SelectedRow()
	if row == nil {
		return
	}
	switch v := row.Data.(type) {
	case ChatItem:
		if !v.IsDivider {
			m.drillDetail(v)
		}
	case ToolCallRow:
		m.SelectedToolCall = &v
		m.drillInto(model.ResourceToolCallDetail)
	}
}

// ApplyExpansion inserts ToolCallRow sub-rows after each expanded ChatItem row
// in the current Table and adds [+]/[-] indicators to expandable rows.
// It preserves the cursor position by key.
func (m *AppModel) ApplyExpansion() {
	// Record the currently selected row's identity for cursor restoration.
	var anchorCIKey string
	var anchorTCName string
	var anchorIsTCRow bool
	if row := m.Table.SelectedRow(); row != nil {
		switch v := row.Data.(type) {
		case ChatItem:
			anchorCIKey = ChatItemKey(v)
		case ToolCallRow:
			anchorCIKey = v.ChatItemKey
			anchorTCName = v.ToolCall.Name
			anchorIsTCRow = true
		}
	}

	// Build expanded row slice, adding [+]/[-] indicators and inserting sub-rows.
	var expanded []Row
	for _, row := range m.Table.Rows {
		ci, isChatItem := row.Data.(ChatItem)
		if isChatItem && !ci.IsDivider && len(ci.AllToolCalls()) > 0 {
			key := ChatItemKey(ci)
			// Copy cells to avoid mutating the underlying array.
			newCells := make([]string, len(row.Cells))
			copy(newCells, row.Cells)
			row.Cells = newCells
			if m.ExpandedItems[key] {
				row.Cells[0] += " [-]"
				expanded = append(expanded, row)
				// Insert sub-rows for tool calls in the primary turn.
				for _, tc := range ci.Turn.ToolCalls {
					expanded = append(expanded, buildToolCallSubRow(ToolCallRow{
						ToolCall: tc, ParentTurn: ci.Turn, ChatItemKey: key,
					}))
				}
				// Insert sub-rows for tool calls in each ExtraTurn.
				for _, et := range ci.ExtraTurns {
					for _, tc := range et.ToolCalls {
						expanded = append(expanded, buildToolCallSubRow(ToolCallRow{
							ToolCall: tc, ParentTurn: et, ChatItemKey: key,
						}))
					}
				}
			} else {
				row.Cells[0] += " [+]"
				expanded = append(expanded, row)
			}
		} else {
			expanded = append(expanded, row)
		}
	}

	m.Table.SetRows(expanded)

	// Restore cursor to the anchor row by key.
	if anchorCIKey == "" {
		return
	}
	rows := m.Table.FilteredRows()
	for i, row := range rows {
		switch v := row.Data.(type) {
		case ChatItem:
			if !anchorIsTCRow && ChatItemKey(v) == anchorCIKey {
				m.Table.Selected = i
				m.Table.EnsureVisible()
				return
			}
		case ToolCallRow:
			if anchorIsTCRow && v.ChatItemKey == anchorCIKey && v.ToolCall.Name == anchorTCName {
				m.Table.Selected = i
				m.Table.EnsureVisible()
				return
			}
		}
	}
}

// buildToolCallSubRow constructs a table Row for a ToolCallRow sub-entry.
func buildToolCallSubRow(tr ToolCallRow) Row {
	tc := tr.ToolCall
	displayName := tc.Name
	if tc.Name == "Agent" || tc.Name == "Task" {
		if agentType := extractStringField(tc, "subagent_type"); agentType != "" {
			displayName = agentDisplayName(model.AgentType(agentType))
		}
	}
	name := "  ▸ " + displayName

	msg := tc.InputSummary()

	var action string
	if tc.IsError {
		action = "✗"
	} else {
		action = "✓"
	}

	var modelTok string
	if m := model.ShortModelName(tr.ParentTurn.ModelName); m != "" {
		tok := tr.ParentTurn.InputTokens + tr.ParentTurn.OutputTokens
		if tok > 0 {
			modelTok = m + ":" + model.FormatTokenCount(tok)
		} else {
			modelTok = m
		}
	}

	var dur string
	if tc.Duration > 0 {
		dur = tc.DurationString()
	}

	return Row{
		Cells: []string{name, msg, action, modelTok, dur},
		Data:  tr,
	}
}

func (m *AppModel) drillInto(rt model.ResourceType) {
	m.filterStack = append(m.filterStack, m.Table.Filter)
	m.Table.Filter = ""
	m.Resource = rt
	m.Crumbs.Push(string(rt))
	m.Filter.Deactivate()
	m.Filter.Input = ""
	m.ContentOffset = 0
	if rt == model.ResourceHistory {
		m.ChatFollow = true
	}
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
	chrome := m.Info.Height(len(m.Menu.NavItems), len(m.Menu.ActionItems), len(m.Menu.UtilItems)) + // info panel (top)
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
	m.Info.Resource = m.Resource // ensure jump-hint guard uses current resource
	infoStr := m.Info.ViewWithMenu(m.Menu)

	// --- 2. Resource title bar ---
	titleStr := m.renderTitleBar()

	// --- 3. Content ---
	var contentStr string
	limit := m.contentHeight()
	switch m.Resource {
	case model.ResourcePluginItemDetail:
		contentStr = RenderPluginItemDetail(m.SelectedPluginItem, m.contentWidth())
	case model.ResourceMemoryDetail:
		contentStr = RenderMemoryDetail(m.SelectedMemory, m.contentWidth())
	case model.ResourceHistoryDetail:
		contentStr = RenderChatItemDetail(m.ChatItems, m.SelectedChatItem, m.contentWidth())
	case model.ResourceToolCallDetail:
		contentStr = RenderToolCallDetail(m.SelectedToolCall, m.contentWidth())
	default:
		contentStr = m.Table.View()
	}
	rawLines := strings.Split(strings.TrimRight(contentStr, "\n"), "\n")
	// For content-only views, apply scroll offset (capped to actual max).
	if isContentView(m.Resource) {
		maxOffset := len(rawLines) - limit
		if maxOffset < 0 {
			maxOffset = 0
		}
		offset := m.ContentOffset
		if offset > maxOffset {
			offset = maxOffset
		}
		if offset > 0 {
			rawLines = rawLines[offset:]
		}
	}
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
