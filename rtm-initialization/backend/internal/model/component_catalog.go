package model

// ComponentCatalogItem 组件目录项（内置组件元信息）
type ComponentCatalogItem struct {
	Type           ComponentType   `json:"type"`
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	DefaultImage   string          `json:"default_image"`
	DefaultVersion string          `json:"default_version"`
	DefaultPort    string          `json:"default_port"`
	EmbeddedImage  string          `json:"embedded_image"` // tar 文件名
	DefaultConfig  ComponentConfig `json:"default_config,omitempty"`
}

// ComponentCatalogItemWithStatus 组件目录项（含实时状态）
type ComponentCatalogItemWithStatus struct {
	ComponentCatalogItem
	Installed       bool            `json:"installed"`
	ComponentID     string          `json:"component_id,omitempty"`
	Status          ComponentStatus `json:"status,omitempty"`
	ContainerState  string          `json:"container_state,omitempty"`
	ContainerID     string          `json:"container_id,omitempty"`
	DockerAvailable bool            `json:"docker_available"`
	DockerError     string          `json:"docker_error,omitempty"`
}
