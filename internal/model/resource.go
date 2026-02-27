package model

// ResourceType identifies the type of resource being viewed.
type ResourceType string

const (
	ResourceProjects ResourceType = "projects"
	ResourceSessions ResourceType = "sessions"
	ResourceAgents   ResourceType = "agents"
	ResourceTools    ResourceType = "tools"
	ResourceTasks    ResourceType = "tasks"
	ResourcePlugins  ResourceType = "plugins"
	ResourceMCP      ResourceType = "mcp"
	ResourceMemory   ResourceType = "memories"
)

// ResourceAliases maps shorthand commands to full resource names.
var ResourceAliases = map[string]ResourceType{
	"p":       ResourceProjects,
	"project": ResourceProjects,
	"s":       ResourceSessions,
	"session": ResourceSessions,
	"a":       ResourceAgents,
	"agent":   ResourceAgents,
	"t":       ResourceTools,
	"tool":    ResourceTools,
	"tk":      ResourceTasks,
	"task":    ResourceTasks,
	"pl":      ResourcePlugins,
	"plugin":  ResourcePlugins,
	"m":       ResourceMCP,
}

// AllResourceNames returns all resource type names for autocomplete.
func AllResourceNames() []string {
	return []string{
		string(ResourceProjects), "p",
		string(ResourceSessions), "s",
		string(ResourceAgents), "a",
		string(ResourceTools), "t",
		string(ResourceTasks), "tk",
		string(ResourcePlugins), "pl",
		string(ResourceMCP), "m",
	}
}

// ResolveResource resolves an alias or full name to a ResourceType.
func ResolveResource(name string) (ResourceType, bool) {
	if alias, ok := ResourceAliases[name]; ok {
		return alias, true
	}
	switch ResourceType(name) {
	case ResourceProjects, ResourceSessions, ResourceAgents,
		ResourceTools, ResourceTasks, ResourcePlugins, ResourceMCP, ResourceMemory:
		return ResourceType(name), true
	}
	return "", false
}
