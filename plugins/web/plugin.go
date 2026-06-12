package web

import "github.com/filipgorny/agent/core"

// WebPlugin provides web fetching skills.
type WebPlugin struct{}

func (WebPlugin) Name() string {
	return "web"
}

func (WebPlugin) Description() string {
	return "Fetch and read web pages."
}

func (WebPlugin) Skills() map[string]func(core.Deps) core.Skill {
	return map[string]func(core.Deps) core.Skill{
		WebGetSkillName:      func(core.Deps) core.Skill { return WebGet{} },
		WebDownloadSkillName: func(core.Deps) core.Skill { return WebDownload{} },
	}
}

func (WebPlugin) Events() []core.EventSpec {
	var events []core.EventSpec

	events = append(events, WebGet{}.GetEvents()...)
	events = append(events, WebDownload{}.GetEvents()...)

	return events
}
