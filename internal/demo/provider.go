package demo

import (
	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

// Provider implements ui.DataProvider with synthetic demo data.
type Provider struct {
	projects []*model.Project
	plugins  []*model.Plugin
}

// NewProvider creates a new demo Provider.
func NewProvider() ui.DataProvider {
	return &Provider{
		projects: GenerateProjects(),
		plugins:  GeneratePlugins(),
	}
}

func (d *Provider) GetProjects() []*model.Project { return d.projects }

func (d *Provider) GetSessions(projectHash string) []*model.Session {
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

func (d *Provider) GetAgents(sessionID string) []*model.Agent {
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

func (d *Provider) GetPlugins(_ string) []*model.Plugin { return d.plugins }

func (d *Provider) GetPluginItems(plugin *model.Plugin) []*model.PluginItem {
	return GeneratePluginItems(plugin.Name)
}

func (d *Provider) GetMemories(_ string) []*model.Memory { return GenerateMemories() }

func (d *Provider) GetTurns(_ string) []model.Turn { return GenerateTurns() }
