package core

// Plugin groups related skills AND the events they emit. Both skills and events
// belong to a plugin. Plugins are registered on an Agent (Agent.RegisterPlugin),
// not via global state; skills are then selected from the registered plugins.
type Plugin interface {
	// Name is the identifier used in config (plugins: list).
	Name() string

	// Description explains, in English, what the plugin provides.
	Description() string

	// Skills returns the plugin's skill factories keyed by skill name.
	Skills() map[string]func(Deps) Skill

	// Events is the catalog of events this plugin defines.
	Events() []EventSpec
}
