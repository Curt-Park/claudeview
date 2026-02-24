package model

// MCPServer represents an MCP server configuration.
type MCPServer struct {
	Name      string
	Plugin    string
	Transport string // stdio, sse, http
	Command   string
	Args      []string
	URL       string
	ToolCount int
	Status    Status
}

// CommandString returns the full command string for display.
func (m *MCPServer) CommandString() string {
	if m.Command == "" {
		return m.URL
	}
	cmd := m.Command
	for _, a := range m.Args {
		cmd += " " + a
	}
	if len(cmd) > 50 {
		return cmd[:49] + "…"
	}
	return cmd
}
