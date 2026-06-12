// Package plugins provides the built-in plugin set.
package plugins

import (
	"github.com/filipgorny/agent/core"
	"github.com/filipgorny/agent/plugins/files"
	"github.com/filipgorny/agent/plugins/interaction"
	"github.com/filipgorny/agent/plugins/reasoning"
	"github.com/filipgorny/agent/plugins/shell"
	"github.com/filipgorny/agent/plugins/text"
	"github.com/filipgorny/agent/plugins/web"
)

// DefaultPlugins returns the built-in plugins. Pass them (or a custom subset) to
// agent.NewAgentFromConfig, or register them with Agent.RegisterPlugin.
func DefaultPlugins() []core.Plugin {
	return []core.Plugin{
		files.FilesPlugin{},
		web.WebPlugin{},
		text.TextPlugin{},
		shell.ShellPlugin{},
		reasoning.ReasoningPlugin{},
		interaction.InteractionPlugin{},
	}
}
