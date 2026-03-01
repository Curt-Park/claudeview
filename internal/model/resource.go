package model

// ResourceType identifies the type of resource being viewed.
type ResourceType string

const (
	ResourceProjects         ResourceType = "projects"
	ResourceSessions         ResourceType = "sessions"
	ResourceAgents           ResourceType = "agents"
	ResourcePlugins          ResourceType = "plugins"
	ResourceMemory           ResourceType = "memories"
	ResourcePluginDetail     ResourceType = "plugin-detail"
	ResourcePluginItemDetail ResourceType = "plugin-item-detail"
	ResourceMemoryDetail     ResourceType = "memory-detail"
)

// ResourceAliases maps shorthand commands to full resource names.
var ResourceAliases = map[string]ResourceType{
	"p":       ResourceProjects,
	"project": ResourceProjects,
	"s":       ResourceSessions,
	"session": ResourceSessions,
	"a":       ResourceAgents,
	"agent":   ResourceAgents,
	"pl":      ResourcePlugins,
	"plugin":  ResourcePlugins,
}

// AllResourceNames returns all resource type names for autocomplete.
func AllResourceNames() []string {
	return []string{
		string(ResourceProjects), "p",
		string(ResourceSessions), "s",
		string(ResourceAgents), "a",
		string(ResourcePlugins), "pl",
		string(ResourceMemory),
	}
}

// ResolveResource resolves an alias or full name to a ResourceType.
func ResolveResource(name string) (ResourceType, bool) {
	if alias, ok := ResourceAliases[name]; ok {
		return alias, true
	}
	switch ResourceType(name) {
	case ResourceProjects, ResourceSessions, ResourceAgents,
		ResourcePlugins, ResourceMemory:
		return ResourceType(name), true
	}
	return "", false
}
