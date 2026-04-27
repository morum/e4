package lobby

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Up         key.Binding
	Down       key.Binding
	Join       key.Binding
	Watch      key.Binding
	Create     key.Binding
	Refresh    key.Binding
	CycleTheme key.Binding
	Help       key.Binding
	Quit       key.Binding
	Cancel     key.Binding
	Submit     key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up:         key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:       key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Join:       key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "join")),
		Watch:      key.NewBinding(key.WithKeys("w"), key.WithHelp("w", "watch")),
		Create:     key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "create")),
		Refresh:    key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
		CycleTheme: key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "theme")),
		Help:       key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		Quit:       key.NewBinding(key.WithKeys("ctrl+c", "q"), key.WithHelp("q", "quit")),
		Cancel:     key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
		Submit:     key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "submit")),
	}
}
