package plugin

// PluginMetadata identifies a plugin to the host CLI.
type PluginMetadata struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	Description string `json:"description"`
	Author     string `json:"author"`
	MinHostVer string `json:"min_host_version"`
}

// PluginInterface is the contract between Cantool and any plugin.
type PluginInterface interface {
	Metadata() PluginMetadata
	Commands() []*PluginCommand
	Hooks() []HookRegistration
}

// PluginCommand describes a command contributed by a plugin.
type PluginCommand struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Usage       string `json:"usage"`
	Flags       []Flag `json:"flags"`
	Args        []Arg  `json:"args"`
}

// Flag describes a command flag.
type Flag struct {
	Name        string `json:"name"`
	Short       string `json:"short,omitempty"`
	Description string `json:"description"`
	Default     string `json:"default,omitempty"`
	Required    bool   `json:"required"`
}

// Arg describes a positional argument.
type Arg struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

// HookPoint defines when in the lifecycle a hook fires.
type HookPoint string

const (
	HookPreBuild   HookPoint = "pre-build"
	HookPostBuild  HookPoint = "post-build"
	HookPreTest    HookPoint = "pre-test"
	HookPostTest   HookPoint = "post-test"
	HookPreDeploy  HookPoint = "pre-deploy"
	HookPostDeploy HookPoint = "post-deploy"
	HookPreDev     HookPoint = "pre-dev"
)

// HookRegistration binds a plugin to a lifecycle event.
type HookRegistration struct {
	Point    HookPoint `json:"point"`
	Priority int       `json:"priority"`
}
