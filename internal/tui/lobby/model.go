package lobby

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/morum/e4/internal/domain"
	"github.com/morum/e4/internal/service"
	"github.com/morum/e4/internal/tui/theme"
	"github.com/morum/e4/internal/tui/widget"
)

const refreshInterval = 2 * time.Second

type mode int

const (
	modeBrowse mode = iota
	modeCreate
)

type Model struct {
	participant domain.Participant
	lobby       *service.LobbyService

	width  int
	height int

	keys      KeyMap
	list      list.Model
	create    textinput.Model
	mode      mode
	helpOn    bool
	statusMsg string
	statusErr bool

	rooms []domain.RoomSummary
}

func New(participant domain.Participant, lobby *service.LobbyService) Model {
	delegate := list.NewDefaultDelegate()
	delegate.SetSpacing(0)
	delegate.ShowDescription = true

	l := list.New(nil, delegate, 60, 14)
	l.Title = "rooms"
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowFilter(false)
	l.SetShowPagination(true)

	create := textinput.New()
	create.Placeholder = "time control like 10|0 (minutes|increment)"
	create.Prompt = "tc › "
	create.CharLimit = 16
	create.Width = 32

	return Model{
		participant: participant,
		lobby:       lobby,
		keys:        DefaultKeyMap(),
		list:        l,
		create:      create,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.refreshCmd(), refreshTicker())
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		listW, listH := m.listSize()
		m.list.SetSize(listW, listH)
		m.create.Width = max(20, msg.Width/3)

	case refreshTickMsg:
		cmds = append(cmds, m.refreshCmd(), refreshTicker())

	case roomsLoadedMsg:
		m.rooms = msg.Rooms
		items := make([]list.Item, len(msg.Rooms))
		for i, r := range msg.Rooms {
			items[i] = roomItem{summary: r}
		}
		cmds = append(cmds, m.list.SetItems(items))

	case roomsErrorMsg:
		m.setError(msg.Err.Error())

	case joinResultMsg:
		if msg.Err != nil {
			m.setError(msg.Err.Error())
		} else {
			return m, func() tea.Msg {
				return EnterRoomMsg{Room: msg.Room, Role: msg.Role, Sub: msg.Sub}
			}
		}

	case tea.KeyMsg:
		switch m.mode {
		case modeCreate:
			return m.updateCreate(msg)
		default:
			return m.updateBrowse(msg)
		}
	}

	if m.mode == modeBrowse {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

func (m Model) updateBrowse(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit
	case key.Matches(msg, m.keys.Refresh):
		return m, m.refreshCmd()
	case key.Matches(msg, m.keys.CycleTheme):
		return m, func() tea.Msg { return CycleThemeMsg{} }
	case key.Matches(msg, m.keys.Help):
		m.helpOn = !m.helpOn
		return m, nil
	case key.Matches(msg, m.keys.Create):
		m.mode = modeCreate
		m.create.SetValue("10|0")
		m.create.Focus()
		m.clearStatus()
		return m, textinput.Blink
	case key.Matches(msg, m.keys.Watch):
		summary, ok := m.selectedSummary()
		if !ok {
			m.setError("no room selected")
			return m, nil
		}
		return m, m.watchCmd(summary.ID)
	case key.Matches(msg, m.keys.Join):
		summary, ok := m.selectedSummary()
		if !ok {
			m.setError("no room selected")
			return m, nil
		}
		if !summary.HasOpenSeat {
			return m, m.watchCmd(summary.ID)
		}
		return m, m.joinCmd(summary.ID)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) updateCreate(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Cancel):
		m.mode = modeBrowse
		m.create.Blur()
		return m, nil
	case key.Matches(msg, m.keys.Submit):
		raw := strings.TrimSpace(m.create.Value())
		tc, err := domain.ParseTimeControl(raw)
		if err != nil {
			m.setError(err.Error())
			return m, nil
		}
		m.mode = modeBrowse
		m.create.Blur()
		return m, m.createCmd(tc)
	}
	var cmd tea.Cmd
	m.create, cmd = m.create.Update(msg)
	return m, cmd
}

func (m Model) View(t theme.Theme) string {
	if m.width == 0 || m.height == 0 {
		return t.Dim.Render("loading lobby…")
	}

	header := m.renderHeader(t)
	body := m.renderBody(t)
	footer := m.renderFooter(t)

	parts := []string{header, "", body, "", footer}
	full := lipgloss.JoinVertical(lipgloss.Left, parts...)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, full)
}

func (m Model) renderHeader(t theme.Theme) string {
	banner := widget.Banner(t)
	hello := t.Muted.Render("connected as ") + t.Accent.Render(m.participant.Nickname)
	counts := t.Dim.Render(m.summaryCounts())
	return strings.Join([]string{banner, "", hello + "  " + counts}, "\n")
}

func (m Model) renderBody(t theme.Theme) string {
	if m.mode == modeCreate {
		title := t.Section.Render("Create room")
		hint := t.Dim.Render("format: <minutes>|<increment-seconds>   examples: 5|0  10|3  15|10")
		return strings.Join([]string{title, "", m.create.View(), "", hint}, "\n")
	}
	return m.list.View()
}

func (m Model) renderFooter(t theme.Theme) string {
	var status string
	if m.statusMsg != "" {
		if m.statusErr {
			status = t.StatusErr.Render("[ERR] ") + m.statusMsg
		} else {
			status = t.StatusInfo.Render("[INFO] ") + m.statusMsg
		}
	}

	var tokens []string
	if m.mode == modeCreate {
		tokens = []string{"enter submit", "esc cancel"}
	} else {
		tokens = []string{
			"↑/↓ select", "enter join", "w watch", "c create",
			"r refresh", "t theme", "? help", "q quit",
		}
	}
	hint := t.Dim.Render(widget.WrapPieces(tokens, "  ", m.width))
	return strings.TrimRight(strings.Join([]string{status, hint}, "\n"), "\n")
}

func (m Model) summaryCounts() string {
	open, active, finished := 0, 0, 0
	for _, r := range m.rooms {
		switch r.Status {
		case domain.RoomStatusActive:
			active++
		case domain.RoomStatusWaiting:
			if r.HasOpenSeat {
				open++
			} else {
				active++
			}
		case domain.RoomStatusFinished:
			finished++
		}
	}
	return fmt.Sprintf("%d open · %d active · %d finished", open, active, finished)
}

func (m Model) selectedSummary() (domain.RoomSummary, bool) {
	item, ok := m.list.SelectedItem().(roomItem)
	if !ok {
		return domain.RoomSummary{}, false
	}
	return item.summary, true
}

func (m Model) listSize() (int, int) {
	// Never let the list exceed the viewport; on narrow terminals it's
	// better to let the list clip its own items than to overflow the edge.
	w := m.width - 4
	if w < 1 {
		w = m.width
	}
	if w > m.width {
		w = m.width
	}
	h := m.height - 14
	if h < 6 {
		h = 6
	}
	return w, h
}

func (m *Model) setError(s string) {
	m.statusMsg = s
	m.statusErr = true
}

func (m *Model) clearStatus() {
	m.statusMsg = ""
	m.statusErr = false
}

func (m Model) refreshCmd() tea.Cmd {
	return func() tea.Msg {
		return roomsLoadedMsg{Rooms: m.lobby.ListGames()}
	}
}

func (m Model) joinCmd(id string) tea.Cmd {
	return func() tea.Msg {
		room, role, err := m.lobby.JoinGame(id, m.participant)
		if err != nil {
			return joinResultMsg{Err: err}
		}
		sub := room.Subscribe()
		return joinResultMsg{Room: room, Role: role, Sub: sub}
	}
}

func (m Model) watchCmd(id string) tea.Cmd {
	return func() tea.Msg {
		room, err := m.lobby.WatchGame(id, m.participant)
		if err != nil {
			return joinResultMsg{Err: err}
		}
		sub := room.Subscribe()
		return joinResultMsg{Room: room, Role: domain.RoleWatcher, Sub: sub}
	}
}

func (m Model) createCmd(tc domain.TimeControl) tea.Cmd {
	return func() tea.Msg {
		room, role, err := m.lobby.CreateGame(m.participant, tc)
		if err != nil {
			return joinResultMsg{Err: err}
		}
		sub := room.Subscribe()
		return joinResultMsg{Room: room, Role: role, Sub: sub}
	}
}

func refreshTicker() tea.Cmd {
	return tea.Tick(refreshInterval, func(time.Time) tea.Msg { return refreshTickMsg{} })
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
