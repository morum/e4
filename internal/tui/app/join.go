package app

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/morum/e4/internal/tui/theme"
	"github.com/morum/e4/internal/tui/widget"
)

type joinModel struct {
	input  textinput.Model
	width  int
	height int
	error  string
}

type nicknameSubmittedMsg struct {
	Nickname string
}

func newJoinModel(suggested string) joinModel {
	ti := textinput.New()
	ti.Placeholder = "pick a nickname"
	ti.Prompt = "› "
	ti.CharLimit = 24
	ti.Width = 28
	ti.SetValue(suggested)
	ti.Focus()
	return joinModel{input: ti}
}

func (j joinModel) Init() tea.Cmd {
	return textinput.Blink
}

func (j joinModel) Update(msg tea.Msg) (joinModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		j.width = msg.Width
		j.height = msg.Height
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))):
			return j, tea.Quit
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			name := strings.TrimSpace(j.input.Value())
			if name == "" {
				j.error = "nickname cannot be empty"
				return j, nil
			}
			return j, func() tea.Msg { return nicknameSubmittedMsg{Nickname: name} }
		}
	}
	var cmd tea.Cmd
	j.input, cmd = j.input.Update(msg)
	return j, cmd
}

func (j joinModel) View(t theme.Theme) string {
	if j.width == 0 || j.height == 0 {
		return ""
	}
	banner := widget.Banner(t)
	prompt := t.Section.Render("Pick a nickname")
	help := t.Dim.Render("enter to continue · ctrl+c to quit")

	parts := []string{banner, "", prompt, "", j.input.View()}
	if j.error != "" {
		parts = append(parts, "", t.StatusErr.Render("[ERR] ")+j.error)
	}
	parts = append(parts, "", help)

	body := lipgloss.JoinVertical(lipgloss.Left, parts...)
	return lipgloss.Place(j.width, j.height, lipgloss.Center, lipgloss.Center, body)
}
