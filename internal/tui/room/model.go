package room

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/morum/e4/internal/domain"
	"github.com/morum/e4/internal/service"
	"github.com/morum/e4/internal/tui/theme"
	"github.com/morum/e4/internal/tui/widget"
)

type Model struct {
	participant domain.Participant
	role        domain.Role
	flipped     bool
	room        service.GameRoom
	updates     <-chan domain.GameSnapshot
	cancelSub   func()

	snapshot domain.GameSnapshot
	hasSnap  bool

	width  int
	height int

	keys      KeyMap
	input     textinput.Model
	inputOn   bool
	helpOn    bool
	statusMsg string
	statusErr bool
}

func New(participant domain.Participant, role domain.Role, room service.GameRoom, sub service.RoomSubscription) Model {
	ti := textinput.New()
	ti.Placeholder = "type a move (e4, Nf3, O-O) — :resign :leave :flip :theme <name>"
	ti.Prompt = "› "
	ti.CharLimit = 64
	ti.Width = 40

	m := Model{
		participant: participant,
		role:        role,
		room:        room,
		updates:     sub.Updates,
		cancelSub:   sub.Cancel,
		keys:        DefaultKeyMap(),
		input:       ti,
	}
	if role != domain.RoleWatcher {
		m.input.Focus()
		m.inputOn = true
	}
	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		waitSnapshot(m.updates),
		tickEvery(),
		textinput.Blink,
	)
}

func (m Model) Cancel() {
	if m.cancelSub != nil {
		m.cancelSub()
	}
}

func (m Model) RoomID() string {
	if m.room != nil {
		return m.room.ID()
	}
	return m.snapshot.RoomID
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.input.Width = max(20, msg.Width-12)

	case SnapshotMsg:
		m.snapshot = domain.GameSnapshot(msg)
		m.hasSnap = true
		cmds = append(cmds, waitSnapshot(m.updates))

	case subscriptionClosedMsg:
		return m, func() tea.Msg { return LeaveRequestMsg{Reason: "room closed"} }

	case tickMsg:
		cmds = append(cmds, tickEvery())

	case flipBoardMsg:
		m.flipped = !m.flipped

	case moveSubmittedMsg:
		if msg.Err != nil {
			m.setError(msg.Err.Error())
		} else {
			m.input.Reset()
			m.clearStatus()
		}

	case resignedMsg:
		if msg.Err != nil {
			m.setError(msg.Err.Error())
		} else {
			m.setInfo("you resigned")
		}

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, func() tea.Msg { return LeaveRequestMsg{Reason: "session ended", Quit: true} }
		case key.Matches(msg, m.keys.Leave):
			return m, func() tea.Msg { return LeaveRequestMsg{Reason: "you left the room"} }
		case key.Matches(msg, m.keys.Resign):
			if m.role == domain.RoleWhite || m.role == domain.RoleBlack {
				return m, m.resignCmd()
			}
		case !m.inputOn && key.Matches(msg, m.keys.CycleTheme):
			return m, func() tea.Msg { return CycleThemeMsg{} }
		case !m.inputOn && key.Matches(msg, m.keys.Flip):
			m.flipped = !m.flipped
			return m, nil
		case !m.inputOn && key.Matches(msg, m.keys.Help):
			m.helpOn = !m.helpOn
			return m, nil
		case key.Matches(msg, m.keys.ToggleFocus) && m.role != domain.RoleWatcher:
			m.toggleFocus()
			return m, nil
		case m.inputOn && key.Matches(msg, m.keys.Submit):
			text := strings.TrimSpace(m.input.Value())
			if text == "" {
				return m, nil
			}
			if cmd := m.handleSubmission(text); cmd != nil {
				return m, cmd
			}
			return m, nil
		}

		if m.inputOn {
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			cmds = append(cmds, cmd)
		}

	default:
		if m.inputOn {
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) toggleFocus() {
	if m.inputOn {
		m.input.Blur()
		m.inputOn = false
	} else {
		m.input.Focus()
		m.inputOn = true
	}
}

func (m *Model) setError(s string) {
	m.statusMsg = s
	m.statusErr = true
}

func (m *Model) setInfo(s string) {
	m.statusMsg = s
	m.statusErr = false
}

func (m *Model) clearStatus() {
	m.statusMsg = ""
	m.statusErr = false
}

func (m Model) handleSubmission(text string) tea.Cmd {
	if strings.HasPrefix(text, ":") {
		return m.handleSlashCommand(text)
	}
	return m.submitMoveCmd(text)
}

func (m Model) handleSlashCommand(text string) tea.Cmd {
	parts := strings.Fields(strings.TrimPrefix(text, ":"))
	if len(parts) == 0 {
		return nil
	}
	switch strings.ToLower(parts[0]) {
	case "leave", "quit":
		return func() tea.Msg { return LeaveRequestMsg{Reason: "you left the room"} }
	case "resign":
		return m.resignCmd()
	case "flip":
		return func() tea.Msg { return flipBoardMsg{} }
	case "theme":
		if len(parts) >= 2 {
			return func() tea.Msg { return SetThemeMsg{Name: parts[1]} }
		}
		return func() tea.Msg { return CycleThemeMsg{} }
	default:
		return func() tea.Msg {
			return moveSubmittedMsg{Err: fmt.Errorf("unknown command :%s", parts[0])}
		}
	}
}

func (m Model) submitMoveCmd(move string) tea.Cmd {
	return func() tea.Msg {
		err := m.room.SubmitMove(m.participant.ID, move)
		return moveSubmittedMsg{Move: move, Err: err}
	}
}

func (m Model) resignCmd() tea.Cmd {
	return func() tea.Msg {
		err := m.room.Resign(m.participant.ID)
		return resignedMsg{Err: err}
	}
}

func (m Model) View(t theme.Theme) string {
	if !m.hasSnap || m.width == 0 || m.height == 0 {
		return t.Dim.Render("loading room…")
	}

	header := m.renderHeader(t)
	clocks := m.renderClocks(t)
	footer := m.renderFooter(t)
	help := m.renderHelp(t)

	// Budget board height from the measured chrome rather than a fixed constant.
	chromeH := lipgloss.Height(header) + 1 // header + blank
	chromeH += lipgloss.Height(clocks) + 1 // clocks + blank
	chromeH += lipgloss.Height(footer)
	if help != "" {
		chromeH += 1 + lipgloss.Height(help) // blank + help
	}

	boardH := m.height - chromeH - 1 // one blank between clocks and board
	if boardH < 6 {
		boardH = 6
	}

	board, moves := m.boardAndMoves(t, boardH)
	body := board
	if moves != "" {
		body = lipgloss.JoinHorizontal(lipgloss.Top, board, "  ", moves)
	}

	parts := []string{header, "", clocks, "", body, "", footer}
	if help != "" {
		parts = append(parts, "", help)
	}

	full := lipgloss.JoinVertical(lipgloss.Left, parts...)

	// Only vertically center when the content actually fits; otherwise
	// lipgloss.Place clips both ends, hiding the header and footer.
	if lipgloss.Height(full) >= m.height {
		return lipgloss.PlaceHorizontal(m.width, lipgloss.Center, full)
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, full)
}

func (m Model) renderHeader(t theme.Theme) string {
	badge := t.BadgeWaiting
	switch m.snapshot.Status {
	case domain.RoomStatusActive:
		badge = t.BadgeActive
	case domain.RoomStatusFinished:
		badge = t.BadgeFinished
	}
	pieces := []string{
		t.Title.Render("e4"),
		t.Section.Render("ROOM " + m.snapshot.RoomID),
		badge.Render(strings.ToUpper(string(m.snapshot.Status))),
		t.Muted.Render(m.snapshot.TimeControl.String()),
		t.Dim.Render("you: " + m.participant.Nickname + " (" + roleLabel(m.role) + ")"),
	}
	return widget.WrapPieces(pieces, "  ", m.width)
}

func (m Model) renderClocks(t theme.Theme) string {
	whiteSide := widget.ClockSide{
		Label:   "White",
		Name:    m.snapshot.WhiteName,
		Time:    m.snapshot.WhiteTimeLeft,
		Active:  m.snapshot.Status == domain.RoomStatusActive && m.snapshot.Turn == "white",
		InCheck: m.snapshot.Board.CheckSquare != "" && m.snapshot.Turn == "white",
	}
	blackSide := widget.ClockSide{
		Label:   "Black",
		Name:    m.snapshot.BlackName,
		Time:    m.snapshot.BlackTimeLeft,
		Active:  m.snapshot.Status == domain.RoomStatusActive && m.snapshot.Turn == "black",
		InCheck: m.snapshot.Board.CheckSquare != "" && m.snapshot.Turn == "black",
	}

	top, bottom := blackSide, whiteSide
	if m.bottomColor() == "black" {
		top, bottom = whiteSide, blackSide
	}

	width := m.width - 4
	if width < 16 {
		width = 16
	}
	if width > m.width {
		width = m.width
	}
	return widget.ClockPair(t, top, bottom, width)
}

func (m Model) boardAndMoves(t theme.Theme, boardH int) (string, string) {
	if boardH < 6 {
		boardH = 6
	}

	movesWidth := 24
	includeMoves := m.width >= 80
	boardW := m.width - 4
	if includeMoves {
		boardW -= movesWidth + 4
	}
	if boardW < 16 {
		boardW = 16
	}

	size := widget.PickBoardSize(boardW, boardH)
	board := widget.Board(t, m.snapshot.Board, widget.BoardOptions{
		Size:        size,
		Orientation: m.orientation(),
	})
	if !includeMoves {
		return board, ""
	}
	moves := widget.MoveList(t, m.snapshot.Moves, movesWidth, boardH)
	return board, moves
}

func (m Model) renderFooter(t theme.Theme) string {
	var parts []string
	if m.role != domain.RoleWatcher {
		parts = append(parts, m.input.View())
	}
	status := m.statusLine(t)
	if status != "" {
		parts = append(parts, status)
	}
	return strings.Join(parts, "\n")
}

func (m Model) statusLine(t theme.Theme) string {
	if m.statusMsg != "" {
		if m.statusErr {
			return t.StatusErr.Render("[ERR] ") + m.statusMsg
		}
		return t.StatusInfo.Render("[INFO] ") + m.statusMsg
	}
	if m.snapshot.Outcome != "" {
		return t.StatusOK.Render("Result: ") + m.snapshot.Outcome + " by " + m.snapshot.Method
	}
	if m.snapshot.LastEvent != "" {
		return t.Muted.Render(m.snapshot.LastEvent)
	}
	return ""
}

func (m Model) renderHelp(t theme.Theme) string {
	if !m.helpOn {
		tokens := []string{
			"enter: submit", "esc: focus", "ctrl+r: resign", "l: leave",
			"f: flip", "t: theme", "?: help", "ctrl+c: quit",
		}
		return t.Dim.Render(widget.WrapPieces(tokens, "  ", m.width))
	}
	lines := []string{
		t.Section.Render("Keybindings"),
		"  enter      submit move or :command",
		"  esc        toggle input focus",
		"  ctrl+r     resign the game",
		"  l          leave room → return to lobby",
		"  f          flip board (hold focus off)",
		"  t          cycle to next theme",
		"  ?          toggle this help",
		"  ctrl+c     quit session",
		"",
		t.Section.Render("Slash commands (typed into the input)"),
		"  :resign           resign the game",
		"  :leave / :quit    leave the room (canonical)",
		"  :flip             flip the board",
		"  :theme <name>     switch to a named theme",
	}
	for i, line := range lines {
		lines[i] = widget.TruncateLine(line, m.width)
	}
	return strings.Join(lines, "\n")
}

func (m Model) orientation() domain.Role {
	if m.role == domain.RoleBlack {
		if m.flipped {
			return domain.RoleWhite
		}
		return domain.RoleBlack
	}
	if m.flipped {
		return domain.RoleBlack
	}
	return domain.RoleWhite
}

func (m Model) bottomColor() string {
	o := m.orientation()
	if o == domain.RoleBlack {
		return "black"
	}
	return "white"
}

func roleLabel(r domain.Role) string {
	switch r {
	case domain.RoleWhite:
		return "white"
	case domain.RoleBlack:
		return "black"
	case domain.RoleWatcher:
		return "watching"
	default:
		return "lobby"
	}
}

func waitSnapshot(updates <-chan domain.GameSnapshot) tea.Cmd {
	return func() tea.Msg {
		snap, ok := <-updates
		if !ok {
			return subscriptionClosedMsg{}
		}
		return SnapshotMsg(snap)
	}
}

func tickEvery() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

type flipBoardMsg struct{}

// SetThemeMsg requests the parent app set a specific theme by name.
type SetThemeMsg struct{ Name string }
