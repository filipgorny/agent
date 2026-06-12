package shell

import "github.com/filipgorny/agent/core"

// ShellPlugin provides the shell_run skill.
type ShellPlugin struct{}

func (ShellPlugin) Name() string {
	return "shell"
}

func (ShellPlugin) Description() string {
	return "Run commands in the system shell."
}

func (ShellPlugin) Skills() map[string]func(core.Deps) core.Skill {
	return map[string]func(core.Deps) core.Skill{
		SkillName: func(core.Deps) core.Skill { return ShellRun{} },
	}
}

func (ShellPlugin) Events() []core.EventSpec {
	return ShellRun{}.GetEvents()
}
