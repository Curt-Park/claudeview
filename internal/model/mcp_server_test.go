package model_test

import (
	"testing"

	"github.com/Curt-Park/claudeview/internal/model"
)

func TestMCPServerCommandString(t *testing.T) {
	tests := []struct {
		name   string
		server *model.MCPServer
		want   string
	}{
		{
			name:   "no command falls back to URL",
			server: &model.MCPServer{Command: "", URL: "https://mcp.example.com/sse"},
			want:   "https://mcp.example.com/sse",
		},
		{
			name:   "command without args",
			server: &model.MCPServer{Command: "npx", Args: nil},
			want:   "npx",
		},
		{
			name:   "command with args",
			server: &model.MCPServer{Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-filesystem"}},
			want:   "npx -y @modelcontextprotocol/server-filesystem",
		},
		{
			// Full command is "npx -y @modelcontextprotocol/server-filesystem --extra-long-flag-that-pushes-past-fifty"
			// len > 50, so truncate: cmd[:49] + "…" = "npx -y @modelcontextprotocol/server-filesystem --…"
			name: "truncation at 50 chars",
			server: &model.MCPServer{
				Command: "npx",
				Args:    []string{"-y", "@modelcontextprotocol/server-filesystem", "--extra-long-flag-that-pushes-past-fifty"},
			},
			want: "npx -y @modelcontextprotocol/server-filesystem --\u2026",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.server.CommandString()
			if got != tc.want {
				t.Errorf("CommandString() = %q, want %q", got, tc.want)
			}
		})
	}
}
