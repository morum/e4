package room

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Submit      key.Binding
	ToggleFocus key.Binding
	Resign      key.Binding
	Leave       key.Binding
	Flip        key.Binding
	CycleTheme  key.Binding
	Help        key.Binding
	Quit        key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Submit: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "submit"),
		),
		ToggleFocus: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "focus"),
		),
		Resign: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("ctrl+r", "resign"),
		),
		Leave: key.NewBinding(
			key.WithKeys("l"),
			key.WithHelp("l", "leave"),
		),
		Flip: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "flip"),
		),
		CycleTheme: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "theme"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
	}
}
