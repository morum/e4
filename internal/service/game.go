package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/morum/e4/internal/clock"
	"github.com/morum/e4/internal/domain"

	"github.com/notnil/chess"
)

var (
	ErrRoomClosed        = errors.New("room is closed")
	ErrRoomFull          = errors.New("room already has two players")
	ErrGameFinished      = errors.New("game is already finished")
	ErrGameNotActive     = errors.New("game is not active")
	ErrNotYourSeat       = errors.New("you are not seated in this game")
	ErrNotYourTurn       = errors.New("it is not your turn")
	ErrWatcherCannotMove = errors.New("watchers cannot make moves")
)

type GameRoom interface {
	ID() string
	Snapshot() domain.GameSnapshot
	Subscribe() RoomSubscription
	JoinPlayer(participant domain.Participant) (domain.Role, error)
	AddWatcher(participant domain.Participant) error
	Leave(participantID string) bool
	SubmitMove(participantID, move string) error
	Resign(participantID string) error
}

type RoomSubscription struct {
	Updates <-chan domain.GameSnapshot
	Cancel  func()
}

type Room struct {
	id       string
	commands chan any
	done     chan struct{}
}

type roomState struct {
	logger       *slog.Logger
	timeControl  domain.TimeControl
	status       domain.RoomStatus
	game         *chess.Game
	notation     chess.AlgebraicNotation
	clock        clock.State
	white        *domain.Participant
	black        *domain.Participant
	watchers     map[string]domain.Participant
	connected    map[string]bool
	moves        []string
	lastMoveFrom string
	lastMoveTo   string
	outcome      string
	method       string
	lastEvent    string
	closed       bool
	seq          int
	subs         map[int]chan domain.GameSnapshot
	persistence  GamePersistence
	persisted    bool
}

type subscribeReq struct {
	reply chan subscribeRes
}

type subscribeRes struct {
	id      int
	updates <-chan domain.GameSnapshot
}

type unsubscribeReq struct {
	id int
}

type snapshotReq struct {
	reply chan domain.GameSnapshot
}

type joinReq struct {
	participant domain.Participant
	reply       chan joinRes
}

type joinRes struct {
	role domain.Role
	err  error
}

type watchReq struct {
	participant domain.Participant
	reply       chan error
}

type leaveReq struct {
	participantID string
	reply         chan bool
}

type moveReq struct {
	participantID string
	move          string
	reply         chan error
}

type resignReq struct {
	participantID string
	reply         chan error
}

func NewRoom(id string, tc domain.TimeControl, logger *slog.Logger) *Room {
	return newRoom(id, roomState{
		logger:      logger,
		timeControl: tc,
		status:      domain.RoomStatusWaiting,
		game:        chess.NewGame(),
		clock:       clock.New(tc),
		watchers:    make(map[string]domain.Participant),
		connected:   make(map[string]bool),
		subs:        make(map[int]chan domain.GameSnapshot),
		persistence: noopPersistence{},
	})
}

func NewPersistentRoom(id string, tc domain.TimeControl, persistence GamePersistence, logger *slog.Logger) *Room {
	if persistence == nil {
		persistence = noopPersistence{}
	}
	return newRoom(id, roomState{
		logger:      logger,
		timeControl: tc,
		status:      domain.RoomStatusWaiting,
		game:        chess.NewGame(),
		clock:       clock.New(tc),
		watchers:    make(map[string]domain.Participant),
		connected:   make(map[string]bool),
		subs:        make(map[int]chan domain.GameSnapshot),
		persistence: persistence,
	})
}

func RestoreRoom(record PersistedRoom, persistence GamePersistence, logger *slog.Logger) (*Room, error) {
	if persistence == nil {
		persistence = noopPersistence{}
	}
	game := chess.NewGame()
	notation := chess.AlgebraicNotation{}
	var lastFrom, lastTo string
	for _, moveText := range record.Moves {
		move, err := notation.Decode(game.Position(), moveText)
		if err != nil {
			return nil, fmt.Errorf("restore room %s: decode move %q: %w", record.ID, moveText, err)
		}
		if err := game.Move(move); err != nil {
			return nil, fmt.Errorf("restore room %s: apply move %q: %w", record.ID, moveText, err)
		}
		lastFrom = move.S1().String()
		lastTo = move.S2().String()
	}
	state := roomState{
		logger:       logger,
		timeControl:  record.TimeControl,
		status:       record.Status,
		game:         game,
		clock:        clock.Restore(pausedRestoreClock(record)),
		white:        record.White,
		black:        record.Black,
		watchers:     make(map[string]domain.Participant),
		connected:    make(map[string]bool),
		moves:        append([]string(nil), record.Moves...),
		lastMoveFrom: lastFrom,
		lastMoveTo:   lastTo,
		outcome:      record.Outcome,
		method:       record.Method,
		lastEvent:    record.LastEvent,
		subs:         make(map[int]chan domain.GameSnapshot),
		persistence:  persistence,
		persisted:    true,
	}
	return newRoom(record.ID, state), nil
}

func pausedRestoreClock(record PersistedRoom) clock.Snapshot {
	snapshot := record.Clock
	if record.Status == domain.RoomStatusActive {
		snapshot.Running = false
	}
	return snapshot
}

func newRoom(id string, state roomState) *Room {
	if state.logger != nil {
		state.logger = state.logger.With("room_id", id)
	}
	if state.persistence == nil {
		state.persistence = noopPersistence{}
	}
	if state.connected == nil {
		state.connected = make(map[string]bool)
	}
	if state.watchers == nil {
		state.watchers = make(map[string]domain.Participant)
	}
	if state.subs == nil {
		state.subs = make(map[int]chan domain.GameSnapshot)
	}

	r := &Room{
		id:       id,
		commands: make(chan any),
		done:     make(chan struct{}),
	}

	go r.loop(state)

	return r
}

func (r *Room) ID() string {
	return r.id
}

func (r *Room) Snapshot() domain.GameSnapshot {
	return request(r, func(reply chan domain.GameSnapshot) any {
		return snapshotReq{reply: reply}
	}, domain.GameSnapshot{RoomID: r.id, Status: domain.RoomStatusFinished})
}

func (r *Room) Subscribe() RoomSubscription {
	reply := make(chan subscribeRes, 1)
	select {
	case r.commands <- subscribeReq{reply: reply}:
	case <-r.done:
		return closedSubscription()
	}

	var res subscribeRes
	select {
	case res = <-reply:
	case <-r.done:
		return closedSubscription()
	}

	return RoomSubscription{
		Updates: res.updates,
		Cancel: func() {
			select {
			case r.commands <- unsubscribeReq{id: res.id}:
			case <-r.done:
			}
		},
	}
}

func (r *Room) JoinPlayer(participant domain.Participant) (domain.Role, error) {
	res := request(r, func(reply chan joinRes) any {
		return joinReq{participant: participant, reply: reply}
	}, joinRes{role: domain.RoleNone, err: ErrRoomClosed})
	return res.role, res.err
}

func (r *Room) AddWatcher(participant domain.Participant) error {
	return request(r, func(reply chan error) any {
		return watchReq{participant: participant, reply: reply}
	}, ErrRoomClosed)
}

func (r *Room) Leave(participantID string) bool {
	return request(r, func(reply chan bool) any {
		return leaveReq{participantID: participantID, reply: reply}
	}, true)
}

func (r *Room) SubmitMove(participantID, move string) error {
	return request(r, func(reply chan error) any {
		return moveReq{participantID: participantID, move: move, reply: reply}
	}, ErrRoomClosed)
}

func (r *Room) Resign(participantID string) error {
	return request(r, func(reply chan error) any {
		return resignReq{participantID: participantID, reply: reply}
	}, ErrRoomClosed)
}

func request[T any](r *Room, build func(chan T) any, closed T) T {
	reply := make(chan T, 1)
	select {
	case r.commands <- build(reply):
	case <-r.done:
		return closed
	}

	select {
	case res := <-reply:
		return res
	case <-r.done:
		return closed
	}
}

func closedSubscription() RoomSubscription {
	updates := make(chan domain.GameSnapshot)
	close(updates)
	return RoomSubscription{
		Updates: updates,
		Cancel:  func() {},
	}
}

func (r *Room) loop(state roomState) {
	defer close(r.done)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case msg := <-r.commands:
			switch req := msg.(type) {
			case snapshotReq:
				req.reply <- state.snapshot(r.id, time.Now())
			case subscribeReq:
				state.seq++
				updates := make(chan domain.GameSnapshot, 4)
				updates <- state.snapshot(r.id, time.Now())
				state.subs[state.seq] = updates
				req.reply <- subscribeRes{id: state.seq, updates: updates}
			case unsubscribeReq:
				if ch, ok := state.subs[req.id]; ok {
					close(ch)
					delete(state.subs, req.id)
				}
			case joinReq:
				role, err := state.join(req.participant)
				if err == nil {
					err = state.persistRoom(r.id)
				}
				req.reply <- joinRes{role: role, err: err}
				if err == nil {
					state.broadcast(r.id)
				}
			case watchReq:
				err := state.addWatcher(req.participant)
				req.reply <- err
				if err == nil {
					state.broadcast(r.id)
				}
			case leaveReq:
				empty := state.leave(req.participantID)
				_ = state.persistRoom(r.id)
				req.reply <- empty
				if empty {
					state.closeSubscribers()
					return
				}
				state.broadcast(r.id)
			case moveReq:
				err := state.submitMove(req.participantID, req.move)
				if err == nil {
					err = state.persistLatestMove(r.id, req.participantID)
				}
				if err == nil {
					err = state.persistRoom(r.id)
				}
				req.reply <- err
				if err == nil {
					state.broadcast(r.id)
				}
			case resignReq:
				err := state.resign(req.participantID)
				if err == nil {
					err = state.persistRoom(r.id)
				}
				req.reply <- err
				if err == nil {
					state.broadcast(r.id)
				}
			}
		case <-ticker.C:
			if state.status != domain.RoomStatusActive {
				continue
			}

			if color, ok := state.clock.Flagged(time.Now()); ok {
				state.finishByTimeout(color)
				_ = state.persistRoom(r.id)
				state.broadcast(r.id)
				continue
			}

			_ = state.persistRoom(r.id)
			state.broadcast(r.id)
		}
	}
}

func (s *roomState) join(participant domain.Participant) (domain.Role, error) {
	if s.closed {
		return domain.RoleNone, ErrRoomClosed
	}
	if s.status == domain.RoomStatusFinished {
		return domain.RoleNone, ErrGameFinished
	}

	if s.white != nil && s.white.ID == participant.ID {
		s.connected[participant.ID] = true
		s.resumeIfReady()
		return domain.RoleWhite, nil
	}
	if s.black != nil && s.black.ID == participant.ID {
		s.connected[participant.ID] = true
		s.resumeIfReady()
		return domain.RoleBlack, nil
	}

	if s.white == nil {
		copy := participant
		s.white = &copy
		s.connected[participant.ID] = true
		s.lastEvent = fmt.Sprintf("%s created the room as White.", participant.Nickname)
		s.ensureWaiting()
		s.log("player seated", "session_id", participant.ID, "nickname", participant.Nickname, "role", domain.RoleWhite)
		return domain.RoleWhite, nil
	}

	if s.black == nil {
		copy := participant
		s.black = &copy
		s.status = domain.RoomStatusActive
		s.connected[participant.ID] = true
		s.lastEvent = fmt.Sprintf("%s joined as Black. Game started.", participant.Nickname)
		s.clock.Start(chess.White, time.Now())
		s.log("player seated", "session_id", participant.ID, "nickname", participant.Nickname, "role", domain.RoleBlack)
		s.log("game started", "session_id", participant.ID, "nickname", participant.Nickname, "turn", chess.White.String())
		return domain.RoleBlack, nil
	}

	return domain.RoleNone, ErrRoomFull
}

func (s *roomState) addWatcher(participant domain.Participant) error {
	if s.closed {
		return ErrRoomClosed
	}

	s.watchers[participant.ID] = participant
	s.lastEvent = fmt.Sprintf("%s is watching.", participant.Nickname)
	s.log("watcher added", "session_id", participant.ID, "nickname", participant.Nickname)
	return nil
}

func (s *roomState) leave(participantID string) bool {
	s.leaveSeat(&s.white, participantID, chess.White, domain.RoleWhite)
	s.leaveSeat(&s.black, participantID, chess.Black, domain.RoleBlack)

	if watcher, ok := s.watchers[participantID]; ok {
		delete(s.watchers, participantID)
		s.lastEvent = fmt.Sprintf("%s stopped watching.", watcher.Nickname)
		s.log("watcher removed", "session_id", participantID, "nickname", watcher.Nickname)
	}

	return s.participantCount() == 0
}

func (s *roomState) leaveSeat(seat **domain.Participant, participantID string, _ chess.Color, role domain.Role) {
	participant := *seat
	if participant == nil || participant.ID != participantID {
		return
	}

	if s.status == domain.RoomStatusActive {
		s.connected[participantID] = false
		s.clock.Stop(time.Now())
		s.lastEvent = fmt.Sprintf("%s disconnected. Game paused.", participant.Nickname)
		s.log("player disconnected active game", "session_id", participantID, "role", role)
	} else {
		s.lastEvent = fmt.Sprintf("%s left the room.", participant.Nickname)
		s.log("player left waiting room", "session_id", participantID, "role", role)
		*seat = nil
		s.ensureWaiting()
	}

	if s.status == domain.RoomStatusFinished {
		*seat = nil
		delete(s.connected, participantID)
	}
}

func (s *roomState) submitMove(participantID, move string) error {
	if s.status != domain.RoomStatusActive {
		return ErrGameNotActive
	}

	color, nickname, _, err := s.playerByID(participantID)
	if err != nil {
		if _, ok := s.watchers[participantID]; ok {
			return ErrWatcherCannotMove
		}
		return err
	}

	turn := s.game.Position().Turn()
	if color != turn {
		return ErrNotYourTurn
	}

	cleanMove := strings.TrimSpace(move)
	parsedMove, err := s.notation.Decode(s.game.Position(), cleanMove)
	if err != nil {
		return fmt.Errorf("invalid move %q", cleanMove)
	}

	moveText := s.notation.Encode(s.game.Position(), parsedMove)
	if err := s.game.Move(parsedMove); err != nil {
		return err
	}

	s.moves = append(s.moves, moveText)
	s.lastMoveFrom = parsedMove.S1().String()
	s.lastMoveTo = parsedMove.S2().String()
	s.clock.Switch(turn, time.Now())
	s.lastEvent = fmt.Sprintf("%s played %s.", nickname, moveText)
	s.log("move submitted", "session_id", participantID, "nickname", nickname, "move", moveText, "turn_next", s.game.Position().Turn().String())

	if s.game.Outcome() != chess.NoOutcome {
		s.status = domain.RoomStatusFinished
		s.outcome = s.game.Outcome().String()
		s.method = s.game.Method().String()
		s.clock.Stop(time.Now())
		s.lastEvent = fmt.Sprintf("%s finished the game with %s.", nickname, moveText)
		s.log("game finished", "outcome", s.outcome, "method", s.method)
	}

	return nil
}

func (s *roomState) resign(participantID string) error {
	if s.status != domain.RoomStatusActive {
		return ErrGameNotActive
	}

	color, nickname, _, err := s.playerByID(participantID)
	if err != nil {
		return err
	}

	s.finishByResignation(color)
	s.lastEvent = fmt.Sprintf("%s resigned.", nickname)
	s.log("player resigned", "session_id", participantID, "nickname", nickname)
	return nil
}

func (s *roomState) finishByResignation(color chess.Color) {
	s.game.Resign(color)
	s.status = domain.RoomStatusFinished
	s.outcome = s.game.Outcome().String()
	s.method = s.game.Method().String()
	s.clock.Stop(time.Now())
}

func (s *roomState) resumeIfReady() {
	if s.status != domain.RoomStatusActive || s.white == nil || s.black == nil {
		return
	}
	if !s.connected[s.white.ID] || !s.connected[s.black.ID] {
		return
	}
	s.clock.Start(s.game.Position().Turn(), time.Now())
	s.lastEvent = "Both players are connected. Game resumed."
}

func (s *roomState) finishByTimeout(loser chess.Color) {
	s.status = domain.RoomStatusFinished
	s.method = "timeout"
	s.clock.Stop(time.Now())
	if loser == chess.White {
		s.outcome = "0-1"
		s.lastEvent = fmt.Sprintf("%s flagged on time.", safeName(s.white))
	} else {
		s.outcome = "1-0"
		s.lastEvent = fmt.Sprintf("%s flagged on time.", safeName(s.black))
	}
	s.log("game finished", "outcome", s.outcome, "method", s.method, "reason", "timeout")
}

func (s *roomState) participantCount() int {
	count := len(s.watchers)
	if s.white != nil {
		count++
	}
	if s.black != nil {
		count++
	}
	return count
}

func (s *roomState) ensureWaiting() {
	if s.status == domain.RoomStatusFinished {
		return
	}
	s.status = domain.RoomStatusWaiting
}

func (s *roomState) playerByID(participantID string) (chess.Color, string, domain.Role, error) {
	if s.white != nil && s.white.ID == participantID {
		return chess.White, s.white.Nickname, domain.RoleWhite, nil
	}
	if s.black != nil && s.black.ID == participantID {
		return chess.Black, s.black.Nickname, domain.RoleBlack, nil
	}
	return chess.NoColor, "", domain.RoleNone, ErrNotYourSeat
}

func (s *roomState) snapshot(roomID string, now time.Time) domain.GameSnapshot {
	whiteTime, blackTime := s.clock.Snapshot(now)
	snapshot := domain.GameSnapshot{
		RoomID:        roomID,
		Status:        s.status,
		TimeControl:   s.timeControl,
		WatcherCount:  len(s.watchers),
		Turn:          strings.ToLower(s.game.Position().Turn().Name()),
		FEN:           s.game.FEN(),
		Board:         buildBoardState(s.game.Position(), s.lastMoveFrom, s.lastMoveTo),
		Moves:         append([]string(nil), s.moves...),
		WhiteTimeLeft: whiteTime,
		BlackTimeLeft: blackTime,
		Outcome:       s.outcome,
		Method:        s.method,
		LastEvent:     s.lastEvent,
	}

	if s.white != nil {
		snapshot.WhiteID = s.white.ID
		snapshot.WhiteName = s.white.Nickname
	}
	if s.black != nil {
		snapshot.BlackID = s.black.ID
		snapshot.BlackName = s.black.Nickname
	}

	if snapshot.Outcome == "" && s.game.Outcome() != chess.NoOutcome {
		snapshot.Outcome = s.game.Outcome().String()
		snapshot.Method = s.game.Method().String()
	}

	return snapshot
}

func (s *roomState) broadcast(roomID string) {
	snapshot := s.snapshot(roomID, time.Now())
	for _, sub := range s.subs {
		select {
		case sub <- snapshot:
		default:
			select {
			case <-sub:
			default:
			}
			select {
			case sub <- snapshot:
			default:
			}
		}
	}
}

func (s *roomState) persistRoom(roomID string) error {
	if s.persistence == nil {
		return nil
	}
	now := time.Now()
	snapshot := s.snapshot(roomID, now)
	clockSnapshot := s.clock.Export(now)
	if !s.persisted {
		if err := s.persistence.CreateRoom(context.Background(), snapshot, clockSnapshot); err != nil {
			return err
		}
		s.persisted = true
		return nil
	}
	return s.persistence.UpdateRoom(context.Background(), snapshot, clockSnapshot)
}

func (s *roomState) persistLatestMove(roomID, playerID string) error {
	if s.persistence == nil || len(s.moves) == 0 {
		return nil
	}
	now := time.Now()
	white, black := s.clock.Snapshot(now)
	return s.persistence.AppendMove(context.Background(), roomID, PersistedMove{
		Ply:            len(s.moves),
		PlayerID:       playerID,
		SAN:            s.moves[len(s.moves)-1],
		From:           s.lastMoveFrom,
		To:             s.lastMoveTo,
		FENAfter:       s.game.FEN(),
		WhiteRemaining: white,
		BlackRemaining: black,
		PlayedAt:       now,
	})
}

func (s *roomState) closeSubscribers() {
	for id, sub := range s.subs {
		close(sub)
		delete(s.subs, id)
	}
	s.closed = true
}

func safeName(participant *domain.Participant) string {
	if participant == nil {
		return "player"
	}
	return participant.Nickname
}

func (s *roomState) log(msg string, attrs ...any) {
	if s.logger == nil {
		return
	}
	s.logger.Info(msg, attrs...)
}
