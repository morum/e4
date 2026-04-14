package app

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/morum/e4/internal/domain"
	"github.com/morum/e4/internal/service"
	"github.com/morum/e4/internal/tui/lobby"
	"github.com/morum/e4/internal/tui/room"
	"github.com/morum/e4/internal/tui/theme"
)

type Screen int

const (
	ScreenJoin Screen = iota
	ScreenLobby
	ScreenRoom
)

type Model struct {
	screen       Screen
	participant  domain.Participant
	theme        theme.Theme
	registry     *theme.Registry
	lobbyService *service.LobbyService

	width  int
	height int

	join  joinModel
	lobby lobby.Model
	room  *room.Model
}

func New(participant domain.Participant, lobbySvc *service.LobbyService, registry *theme.Registry, defaultTheme theme.Theme) Model {
	return Model{
		screen:       ScreenJoin,
		participant:  participant,
		theme:        defaultTheme,
		registry:     registry,
		lobbyService: lobbySvc,
		join:         newJoinModel(participant.Nickname),
	}
}

func (m Model) Init() tea.Cmd {
	return m.join.Init()
}

// ThemeName returns the name of the theme that will be used on the next
// render. Exported so tests can verify theme cycling without rendering.
func (m Model) ThemeName() string { return m.theme.Name }

// Screen returns the current screen the model is displaying. Exposed for
// tests; callers shouldn't need to branch on this at runtime.
func (m Model) Screen() Screen { return m.screen }

// InRoom reports whether the model currently owns a room handle. Tests
// use this to assert leave/quit cleanup actually dropped the handle.
func (m Model) InRoom() bool { return m.room != nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch v := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = v.Width
		m.height = v.Height

	case nicknameSubmittedMsg:
		m.participant.Nickname = v.Nickname
		m.lobby = lobby.New(m.participant, m.lobbyService)
		m.screen = ScreenLobby
		return m, tea.Batch(
			m.lobby.Init(),
			func() tea.Msg { return tea.WindowSizeMsg{Width: m.width, Height: m.height} },
		)

	case lobby.EnterRoomMsg:
		rm := room.New(m.participant, v.Role, v.Room, v.Sub)
		m.room = &rm
		m.screen = ScreenRoom
		return m, tea.Batch(
			m.room.Init(),
			func() tea.Msg { return tea.WindowSizeMsg{Width: m.width, Height: m.height} },
		)

	case room.LeaveRequestMsg:
		if m.room != nil {
			m.room.Cancel()
			_ = m.lobbyService.LeaveRoom(roomID(m.room), m.participant.ID)
			m.room = nil
		}
		if v.Quit {
			return m, tea.Quit
		}
		m.screen = ScreenLobby
		return m, tea.Batch(
			m.lobby.Init(),
			func() tea.Msg { return tea.WindowSizeMsg{Width: m.width, Height: m.height} },
		)

	case lobby.CycleThemeMsg, room.CycleThemeMsg:
		m.theme = m.registry.Next(m.theme.Name)
		return m, nil

	case lobby.SetThemeMsg:
		m.applyThemeName(v.Name)
		return m, nil

	case room.SetThemeMsg:
		m.applyThemeName(v.Name)
		return m, nil
	}

	return m.dispatch(msg)
}

func (m Model) dispatch(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.screen {
	case ScreenJoin:
		m.join, cmd = m.join.Update(msg)
	case ScreenLobby:
		m.lobby, cmd = m.lobby.Update(msg)
	case ScreenRoom:
		if m.room != nil {
			updated, c := m.room.Update(msg)
			*m.room = updated
			cmd = c
		}
	}
	return m, cmd
}

func (m *Model) applyThemeName(name string) {
	if t, ok := m.registry.Get(strings.TrimSpace(name)); ok {
		m.theme = t
	}
}

func (m Model) View() string {
	switch m.screen {
	case ScreenJoin:
		return m.join.View(m.theme)
	case ScreenLobby:
		return m.lobby.View(m.theme)
	case ScreenRoom:
		if m.room == nil {
			return m.theme.Dim.Render("loading…")
		}
		return m.room.View(m.theme)
	default:
		return ""
	}
}

func roomID(r *room.Model) string {
	return r.RoomID()
}
