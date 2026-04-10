package ssh

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/morum/e4/internal/domain"
	"github.com/morum/e4/internal/render"
	"github.com/morum/e4/internal/service"

	gssh "github.com/gliderlabs/ssh"
	cryptossh "golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

type Config struct {
	ListenAddr  string
	HostKeyPath string
	Logger      *slog.Logger
}

type Server struct {
	lobby  *service.LobbyService
	server *gssh.Server
	logger *slog.Logger
}

func NewServer(cfg Config, lobby *service.LobbyService) (*Server, error) {
	signer, err := loadOrCreateHostKey(cfg.HostKeyPath)
	if err != nil {
		return nil, err
	}

	s := &Server{lobby: lobby, logger: cfg.Logger}
	s.server = &gssh.Server{
		Addr:        cfg.ListenAddr,
		Handler:     s.handleSession,
		IdleTimeout: 8 * time.Hour,
		MaxTimeout:  24 * time.Hour,
		Version:     "SSH-2.0-e4",
		Banner:      "Welcome to e4\n",
	}
	s.server.AddHostKey(signer)

	return s, nil
}

func (s *Server) ListenAndServe(ctx context.Context) error {
	if s.logger != nil {
		s.logger.Info("ssh server starting", "listen_addr", s.server.Addr)
	}
	errCh := make(chan error, 1)
	go func() {
		errCh <- s.server.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if s.logger != nil && err != nil && err != gssh.ErrServerClosed {
			s.logger.Error("ssh server stopped with error", "error", err)
		}
		if err == nil || err == gssh.ErrServerClosed {
			return nil
		}
		return err
	case <-ctx.Done():
		if s.logger != nil {
			s.logger.Info("ssh server shutting down")
		}
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.server.Shutdown(shutdownCtx)
		err := <-errCh
		if err == nil || err == gssh.ErrServerClosed {
			return nil
		}
		return err
	}
}

func (s *Server) handleSession(sess gssh.Session) {
	if s.logger != nil {
		s.logger.Info("session connected", "remote_addr", sess.RemoteAddr().String(), "ssh_user", sess.User())
	}

	if len(sess.Command()) > 0 {
		_, _ = io.WriteString(sess, "Use an interactive shell session for e4.\n")
		if s.logger != nil {
			s.logger.Warn("rejected exec session", "remote_addr", sess.RemoteAddr().String(), "command", sess.RawCommand())
		}
		_ = sess.Exit(1)
		return
	}

	pty, winCh, hasPTY := sess.Pty()
	client := newClientSession(sess, hasPTY, pty.Term, s.lobby, s.logger)
	defer client.close()
	client.logInfo("session ready", "session_id", client.id, "remote_addr", sess.RemoteAddr().String(), "ssh_user", sess.User(), "has_pty", hasPTY)
	if hasPTY {
		client.setSize(pty.Window.Width, pty.Window.Height)
		go client.watchWindows(winCh)
	}

	client.sendScreen(render.JoinBanner(client.ansi))
	nickname, err := client.readLine("nickname> ")
	if err != nil {
		client.logDebug("nickname read failed", "error", err)
		return
	}

	nickname = strings.TrimSpace(nickname)
	if nickname == "" {
		nickname = fmt.Sprintf("guest-%s", client.id[len(client.id)-4:])
	}
	client.nickname = nickname
	client.logInfo("nickname set", "session_id", client.id, "nickname", nickname)
	client.renderLobby()

	for {
		line, err := client.readLine(client.currentPrompt())
		if err != nil {
			if err != io.EOF {
				client.logDebug("session read ended", "error", err)
			}
			break
		}
		line = strings.TrimSpace(line)
		if err := client.handleLine(line); err != nil {
			if err == io.EOF {
				return
			}
			client.logDebug("command failed", "line", line, "error", err)
			client.setStatus(render.StatusError, err.Error())
			client.renderActiveView()
		}
	}

	client.leaveCurrentRoom()
	if s.logger != nil {
		s.logger.Info("session disconnected", "remote_addr", sess.RemoteAddr().String(), "ssh_user", sess.User(), "session_id", client.id, "nickname", client.nickname)
	}
}

type clientSession struct {
	id       string
	nickname string
	sess     gssh.Session
	hasPTY   bool
	ansi     bool
	lobby    *service.LobbyService
	done     chan struct{}
	term     *term.Terminal
	termMu   sync.Mutex
	logger   *slog.Logger

	mu          sync.RWMutex
	width       int
	height      int
	status      render.StatusLine
	roomID      string
	room        service.GameRoom
	role        domain.Role
	snapshot    *domain.GameSnapshot
	unsubscribe func()
	roomToken   uint64
}

func newClientSession(sess gssh.Session, hasPTY bool, termName string, lobby *service.LobbyService, logger *slog.Logger) *clientSession {
	c := &clientSession{
		id:     randomID(),
		sess:   sess,
		hasPTY: hasPTY,
		ansi:   detectANSI(hasPTY, termName),
		lobby:  lobby,
		done:   make(chan struct{}),
		term:   term.NewTerminal(sess, ""),
		logger: logger,
		width:  100,
		height: 32,
	}
	return c
}

func (c *clientSession) close() {
	c.leaveCurrentRoom()
	select {
	case <-c.done:
		return
	default:
		close(c.done)
	}
}

func (c *clientSession) handleLine(line string) error {
	if line == "" {
		return nil
	}
	c.logDebug("command received", "session_id", c.id, "nickname", c.nickname, "line", line)

	fields := strings.Fields(line)
	command := strings.ToLower(fields[0])
	args := fields[1:]

	if c.inRoom() {
		return c.handleRoomLine(command, line)
	}

	return c.handleLobbyLine(command, args)
}

func (c *clientSession) handleRoomLine(command, line string) error {
	switch command {
	case "help":
		c.setStatus(render.StatusInfo, render.HelpText(true))
		c.renderCurrentRoom()
		return nil
	case "board":
		c.clearStatus()
		c.renderCurrentRoom()
		return nil
	case "leave":
		roomID := c.currentRoomID()
		c.leaveCurrentRoom()
		c.setStatus(render.StatusInfo, fmt.Sprintf("Left room %s.", roomID))
		c.renderLobby()
		return nil
	case "resign":
		c.clearStatus()
		return c.resignCurrentGame()
	case "quit", "exit":
		return c.exitSession()
	}

	if c.currentRole() == domain.RoleWatcher {
		return fmt.Errorf("watchers cannot move")
	}

	c.clearStatus()
	return c.playMove(line)
}

func (c *clientSession) handleLobbyLine(command string, args []string) error {
	switch command {
	case "help":
		c.setStatus(render.StatusInfo, render.HelpText(false))
		c.renderLobby()
		return nil
	case "list":
		c.clearStatus()
		c.renderLobby()
		return nil
	case "create":
		if len(args) != 1 {
			return fmt.Errorf("usage: create <minutes>|<increment>")
		}
		return c.createRoom(args[0])
	case "join":
		if len(args) != 1 {
			return fmt.Errorf("usage: join <room-id>")
		}
		return c.joinRoom(strings.ToUpper(args[0]))
	case "watch":
		if len(args) != 1 {
			return fmt.Errorf("usage: watch <room-id>")
		}
		return c.watchRoom(strings.ToUpper(args[0]))
	case "quit", "exit":
		return c.exitSession()
	default:
		return fmt.Errorf("unknown command %q", command)
	}
}

func (c *clientSession) createRoom(rawTC string) error {
	tc, err := domain.ParseTimeControl(rawTC)
	if err != nil {
		return err
	}

	c.clearStatus()
	room, role, err := c.lobby.CreateGame(c.participant(), tc)
	if err != nil {
		return err
	}

	c.setStatus(render.StatusSuccess, fmt.Sprintf("Created room %s as White.", room.ID()))
	c.attachRoom(room, role)
	return nil
}

func (c *clientSession) joinRoom(roomID string) error {
	c.clearStatus()
	room, role, err := c.lobby.JoinGame(roomID, c.participant())
	if err != nil {
		return err
	}

	c.setStatus(render.StatusSuccess, fmt.Sprintf("Joined room %s as %s.", room.ID(), role))
	c.attachRoom(room, role)
	return nil
}

func (c *clientSession) watchRoom(roomID string) error {
	c.clearStatus()
	room, err := c.lobby.WatchGame(roomID, c.participant())
	if err != nil {
		return err
	}

	c.setStatus(render.StatusInfo, fmt.Sprintf("Watching room %s.", room.ID()))
	c.attachRoom(room, domain.RoleWatcher)
	return nil
}

func (c *clientSession) resignCurrentGame() error {
	room := c.currentRoom()
	if room == nil {
		return fmt.Errorf("you are not in a room")
	}
	if c.currentRole() == domain.RoleWatcher {
		return fmt.Errorf("watchers cannot resign")
	}
	if err := room.Resign(c.id); err != nil {
		return err
	}
	return nil
}

func (c *clientSession) playMove(move string) error {
	room := c.currentRoom()
	if room == nil {
		return fmt.Errorf("you are not in a room")
	}
	return room.SubmitMove(c.id, move)
}

func (c *clientSession) attachRoom(room service.GameRoom, role domain.Role) {
	c.leaveCurrentRoom()

	sub := room.Subscribe()
	token := atomic.AddUint64(&c.roomToken, 1)
	initialSnapshot, ok := <-sub.Updates
	if !ok {
		return
	}

	c.mu.Lock()
	c.room = room
	c.roomID = room.ID()
	c.role = role
	c.snapshot = &initialSnapshot
	c.unsubscribe = sub.Cancel
	c.mu.Unlock()
	c.renderCurrentRoom()

	go func(roomID string, role domain.Role, token uint64, updates <-chan domain.GameSnapshot) {
		for snapshot := range updates {
			if atomic.LoadUint64(&c.roomToken) != token {
				return
			}
			c.mu.Lock()
			s := snapshot
			c.snapshot = &s
			ctx := c.renderContextLocked(role)
			c.mu.Unlock()
			c.sendScreen(render.RoomView(ctx, snapshot, c.nickname, role))
		}

		c.mu.Lock()
		if c.roomID == roomID {
			c.roomID = ""
			c.room = nil
			c.role = domain.RoleNone
			c.snapshot = nil
			c.unsubscribe = nil
		}
		c.mu.Unlock()
	}(room.ID(), role, token, sub.Updates)
}

func (c *clientSession) leaveCurrentRoom() {
	c.mu.Lock()
	roomID := c.roomID
	unsubscribe := c.unsubscribe
	c.roomID = ""
	c.room = nil
	c.role = domain.RoleNone
	c.snapshot = nil
	c.unsubscribe = nil
	c.mu.Unlock()

	if unsubscribe != nil {
		atomic.AddUint64(&c.roomToken, 1)
		unsubscribe()
	}

	if roomID != "" {
		_ = c.lobby.LeaveRoom(roomID, c.id)
	}
}

func (c *clientSession) renderLobby() {
	rooms := c.lobby.ListGames()
	c.mu.RLock()
	ctx := c.renderContextLocked(domain.RoleNone)
	c.mu.RUnlock()
	c.sendScreen(render.LobbyView(ctx, c.nickname, rooms))
}

func (c *clientSession) renderCurrentRoom() {
	c.mu.RLock()
	snapshot := c.snapshot
	role := c.role
	ctx := c.renderContextLocked(role)
	c.mu.RUnlock()
	if snapshot == nil {
		c.renderLobby()
		return
	}
	c.sendScreen(render.RoomView(ctx, *snapshot, c.nickname, role))
}

func (c *clientSession) currentRoom() service.GameRoom {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.room
}

func (c *clientSession) inRoom() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.room != nil
}

func (c *clientSession) currentPrompt() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.snapshot == nil && c.roomID != "" {
		placeholder := &domain.GameSnapshot{RoomID: c.roomID}
		return render.Prompt(c.renderContextLocked(c.role), placeholder, c.role)
	}
	return render.Prompt(c.renderContextLocked(c.role), c.snapshot, c.role)
}

func (c *clientSession) currentRole() domain.Role {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.role
}

func (c *clientSession) sendMessage(message string) {
	message = strings.TrimSpace(message)
	if message == "" {
		return
	}
	c.writeTerminal(message + "\n")
}

func (c *clientSession) sendScreen(body string) {
	var b strings.Builder
	if c.hasPTY {
		b.WriteString("\x1b[2J\x1b[H")
	}
	b.WriteString(strings.TrimRight(body, "\n"))
	b.WriteString("\n")
	c.writeTerminal(b.String())
}

func (c *clientSession) writeTerminal(message string) {
	select {
	case <-c.done:
		return
	default:
	}

	c.termMu.Lock()
	defer c.termMu.Unlock()
	_, _ = c.term.Write([]byte(message))
}

func (c *clientSession) readLine(prompt string) (string, error) {
	select {
	case <-c.done:
		return "", io.EOF
	default:
	}

	c.term.SetPrompt(prompt)
	line, err := c.term.ReadLine()
	if err != nil {
		return "", err
	}
	return line, nil
}

func (c *clientSession) setSize(width, height int) bool {
	if width <= 0 || height <= 0 {
		return false
	}
	c.mu.Lock()
	changed := c.width != width || c.height != height
	c.width = width
	c.height = height
	c.mu.Unlock()
	_ = c.term.SetSize(width, height)
	return changed
}

func (c *clientSession) watchWindows(winCh <-chan gssh.Window) {
	for {
		select {
		case <-c.done:
			return
		case win, ok := <-winCh:
			if !ok {
				return
			}
			if c.setSize(win.Width, win.Height) {
				c.renderActiveView()
			}
		}
	}
}

func (c *clientSession) logInfo(msg string, attrs ...any) {
	if c.logger == nil {
		return
	}
	c.logger.Info(msg, attrs...)
}

func (c *clientSession) logDebug(msg string, attrs ...any) {
	if c.logger == nil {
		return
	}
	c.logger.Debug(msg, attrs...)
}

func (c *clientSession) participant() domain.Participant {
	return domain.Participant{ID: c.id, Nickname: c.nickname}
}

func (c *clientSession) currentRoomID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.roomID
}

func (c *clientSession) setStatus(kind render.StatusKind, message string) {
	c.mu.Lock()
	c.status = render.StatusLine{Kind: kind, Message: strings.TrimSpace(message)}
	c.mu.Unlock()
}

func (c *clientSession) clearStatus() {
	c.mu.Lock()
	c.status = render.StatusLine{}
	c.mu.Unlock()
}

func (c *clientSession) renderActiveView() {
	if c.inRoom() {
		c.renderCurrentRoom()
		return
	}
	c.renderLobby()
}

func (c *clientSession) renderContextLocked(role domain.Role) render.Context {
	return render.Context{
		Width:       c.width,
		Height:      c.height,
		ANSI:        c.ansi,
		Role:        role,
		Orientation: render.OrientationForRole(role),
		Status:      c.status,
	}
}

func (c *clientSession) exitSession() error {
	_ = c.sess.Exit(0)
	return io.EOF
}

func detectANSI(hasPTY bool, termName string) bool {
	if !hasPTY {
		return false
	}

	if termName == "" || termName == "dumb" {
		return false
	}

	return true
}

func randomID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("session-%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("%x", buf)
}

func loadOrCreateHostKey(path string) (cryptossh.Signer, error) {
	if pemBytes, err := os.ReadFile(path); err == nil {
		return cryptossh.ParsePrivateKey(pemBytes)
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}
	if err := os.WriteFile(path, pemBytes, 0o600); err != nil {
		return nil, err
	}

	return cryptossh.NewSignerFromKey(key)
}
