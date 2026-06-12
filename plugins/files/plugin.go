package files

import "github.com/filipgorny/agent/core"

// FilesPlugin provides filesystem skills.
type FilesPlugin struct{}

func (FilesPlugin) Name() string {
	return "files"
}

func (FilesPlugin) Description() string {
	return "Read, write, edit, search, list and watch files on the local filesystem."
}

func (FilesPlugin) Skills() map[string]func(core.Deps) core.Skill {
	return map[string]func(core.Deps) core.Skill{
		FileReadSkillName:  func(core.Deps) core.Skill { return FileRead{} },
		FileWriteSkillName: func(core.Deps) core.Skill { return FileWrite{} },
		FileEditSkillName:  func(core.Deps) core.Skill { return FileEdit{} },
		GrepSkillName:      func(core.Deps) core.Skill { return Grep{} },
		DirListSkillName:   func(core.Deps) core.Skill { return DirList{} },
		FileWatchSkillName: func(d core.Deps) core.Skill { return &FileWatch{emit: d.Emit} },
	}
}

func (FilesPlugin) Events() []core.EventSpec {
	var events []core.EventSpec

	events = append(events, FileRead{}.GetEvents()...)
	events = append(events, FileWrite{}.GetEvents()...)
	events = append(events, FileEdit{}.GetEvents()...)
	events = append(events, Grep{}.GetEvents()...)
	events = append(events, DirList{}.GetEvents()...)
	events = append(events, (&FileWatch{}).GetEvents()...)

	return events
}
