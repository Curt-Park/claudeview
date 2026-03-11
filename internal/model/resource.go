package model

// ResourceType identifies the type of resource being viewed.
type ResourceType string

const (
	ResourceProjects         ResourceType = "projects"
	ResourceSessions         ResourceType = "sessions"
	ResourcePlugins          ResourceType = "plugins"
	ResourceMemory           ResourceType = "memories"
	ResourcePluginDetail     ResourceType = "plugin-detail"
	ResourcePluginItemDetail ResourceType = "plugin-item-detail"
	ResourceMemoryDetail     ResourceType = "memory-detail"
	ResourceHistory          ResourceType = "history"
	ResourceHistoryDetail    ResourceType = "history-detail"
	ResourceToolCallDetail   ResourceType = "tool-call-detail"
)
