package postgres

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/morum/e4/internal/clock"
	"github.com/morum/e4/internal/domain"
	"github.com/morum/e4/internal/service"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/notnil/chess"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

type Store struct {
	db *sql.DB
}

func Open(ctx context.Context, databaseURL string) (*Store, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) Migrate(ctx context.Context) error {
	entries, err := migrationFiles.ReadDir("migrations")
	if err != nil {
		return err
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)

	for _, name := range names {
		version := strings.TrimSuffix(name, ".sql")
		applied, err := s.migrationApplied(ctx, version)
		if err != nil {
			return err
		}
		if applied {
			continue
		}
		body, err := migrationFiles.ReadFile("migrations/" + name)
		if err != nil {
			return err
		}
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, string(body)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("migration %s: %w", version, err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations(version) VALUES ($1)`, version); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record migration %s: %w", version, err)
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) migrationApplied(ctx context.Context, version string) (bool, error) {
	var exists bool
	err := s.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'schema_migrations'
		)
	`).Scan(&exists)
	if err != nil || !exists {
		return false, err
	}
	err = s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)`, version).Scan(&exists)
	return exists, err
}

func (s *Store) FindOrCreateBySSHKey(ctx context.Context, key service.SSHKeyIdentity, nickname string) (domain.Participant, error) {
	if strings.TrimSpace(key.Fingerprint) == "" {
		return domain.Participant{}, errors.New("missing SSH key fingerprint")
	}
	nickname = strings.TrimSpace(nickname)
	if nickname == "" {
		nickname = "guest"
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Participant{}, err
	}
	defer tx.Rollback()

	participant, found, err := findParticipantByKey(ctx, tx, key)
	if err != nil {
		return domain.Participant{}, err
	}
	if found {
		return participant, tx.Commit()
	}

	var playerID string
	if err := tx.QueryRowContext(ctx, `INSERT INTO players(nickname) VALUES ($1) RETURNING id`, nickname).Scan(&playerID); err != nil {
		return domain.Participant{}, err
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO player_keys(fingerprint, player_id, authorized_key, key_type)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (fingerprint) DO NOTHING
	`, key.Fingerprint, playerID, key.AuthorizedKey, key.KeyType)
	if err != nil {
		return domain.Participant{}, err
	}
	participant, found, err = findParticipantByKey(ctx, tx, key)
	if err != nil {
		return domain.Participant{}, err
	}
	if !found {
		return domain.Participant{}, errors.New("failed to persist SSH key identity")
	}
	return participant, tx.Commit()
}

type queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func findParticipantByKey(ctx context.Context, q queryer, key service.SSHKeyIdentity) (domain.Participant, bool, error) {
	var p domain.Participant
	err := q.QueryRowContext(ctx, `
		SELECT p.id::text, p.nickname, k.fingerprint
		FROM player_keys k
		JOIN players p ON p.id = k.player_id
		WHERE k.fingerprint = $1
	`, key.Fingerprint).Scan(&p.ID, &p.Nickname, &p.KeyFingerprint)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Participant{}, false, nil
	}
	if err != nil {
		return domain.Participant{}, false, err
	}
	p.PlayerID = p.ID
	return p, true, nil
}

func (s *Store) CreateRoom(ctx context.Context, snapshot domain.GameSnapshot, clockSnapshot clock.Snapshot) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO games (
			room_id, status, base_ns, increment_ns,
			white_player_id, white_name, black_player_id, black_name,
			turn, fen, white_remaining_ns, black_remaining_ns,
			clock_running, active_color, clock_updated_at,
			outcome, method, last_event, finished_at
		)
		VALUES (
			$1, $2, $3, $4,
			CAST($5 AS uuid), $6, CAST($7 AS uuid), $8,
			$9, $10, $11, $12,
			$13, $14, $15,
			$16, $17, $18,
			CASE WHEN $2 = 'finished' THEN now() ELSE NULL END
		)
		ON CONFLICT (room_id) DO NOTHING
	`, roomValues(snapshot, clockSnapshot)...)
	return err
}

func (s *Store) UpdateRoom(ctx context.Context, snapshot domain.GameSnapshot, clockSnapshot clock.Snapshot) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE games SET
			status = $2,
			base_ns = $3,
			increment_ns = $4,
			white_player_id = CAST($5 AS uuid),
			white_name = $6,
			black_player_id = CAST($7 AS uuid),
			black_name = $8,
			turn = $9,
			fen = $10,
			white_remaining_ns = $11,
			black_remaining_ns = $12,
			clock_running = $13,
			active_color = $14,
			clock_updated_at = $15,
			outcome = $16,
			method = $17,
			last_event = $18,
			updated_at = now(),
			finished_at = CASE WHEN $2 = 'finished' AND finished_at IS NULL THEN now() ELSE finished_at END
		WHERE room_id = $1
	`, roomValues(snapshot, clockSnapshot)...)
	return err
}

func roomValues(snapshot domain.GameSnapshot, clockSnapshot clock.Snapshot) []any {
	return []any{
		snapshot.RoomID,
		string(snapshot.Status),
		int64(snapshot.TimeControl.Base),
		int64(snapshot.TimeControl.Increment),
		emptyToNil(snapshot.WhiteID),
		snapshot.WhiteName,
		emptyToNil(snapshot.BlackID),
		snapshot.BlackName,
		snapshot.Turn,
		snapshot.FEN,
		int64(clockSnapshot.WhiteRemaining),
		int64(clockSnapshot.BlackRemaining),
		clockSnapshot.Running,
		colorName(clockSnapshot.ActiveColor),
		clockSnapshot.LastUpdate,
		snapshot.Outcome,
		snapshot.Method,
		snapshot.LastEvent,
	}
}

func (s *Store) AppendMove(ctx context.Context, roomID string, move service.PersistedMove) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO game_moves (
			room_id, ply, player_id, san, from_square, to_square, fen_after,
			white_remaining_ns, black_remaining_ns, played_at
		)
		VALUES ($1, $2, CAST($3 AS uuid), $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (room_id, ply) DO NOTHING
	`, roomID, move.Ply, emptyToNil(move.PlayerID), move.SAN, move.From, move.To, move.FENAfter,
		int64(move.WhiteRemaining), int64(move.BlackRemaining), move.PlayedAt)
	return err
}

func (s *Store) AppendEvent(ctx context.Context, roomID string, event service.PersistedEvent) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO game_events(room_id, event_type, player_id, message, created_at)
		VALUES ($1, $2, CAST($3 AS uuid), $4, $5)
	`, roomID, event.Type, emptyToNil(event.PlayerID), event.Message, event.CreatedAt)
	return err
}

func (s *Store) LoadOpenRooms(ctx context.Context) ([]service.PersistedRoom, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT
			room_id, status, base_ns, increment_ns,
			white_player_id::text, white_name, black_player_id::text, black_name,
			white_remaining_ns, black_remaining_ns, clock_running, active_color, clock_updated_at,
			outcome, method, last_event
		FROM games
		WHERE status = 'active'
			OR (status = 'waiting' AND (white_player_id IS NOT NULL OR black_player_id IS NOT NULL))
		ORDER BY created_at, room_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []service.PersistedRoom
	for rows.Next() {
		var room service.PersistedRoom
		var status string
		var baseNS, incrementNS, whiteNS, blackNS int64
		var whiteID, blackID, whiteName, blackName sql.NullString
		var running bool
		var activeColor string
		var clockUpdated time.Time
		if err := rows.Scan(
			&room.ID, &status, &baseNS, &incrementNS,
			&whiteID, &whiteName, &blackID, &blackName,
			&whiteNS, &blackNS, &running, &activeColor, &clockUpdated,
			&room.Outcome, &room.Method, &room.LastEvent,
		); err != nil {
			return nil, err
		}
		room.Status = domain.RoomStatus(status)
		room.TimeControl = domain.TimeControl{Base: time.Duration(baseNS), Increment: time.Duration(incrementNS)}
		if whiteID.Valid {
			room.White = &domain.Participant{ID: whiteID.String, PlayerID: whiteID.String, Nickname: whiteName.String}
		}
		if blackID.Valid {
			room.Black = &domain.Participant{ID: blackID.String, PlayerID: blackID.String, Nickname: blackName.String}
		}
		room.Clock = clock.Snapshot{
			WhiteRemaining: time.Duration(whiteNS),
			BlackRemaining: time.Duration(blackNS),
			Increment:      time.Duration(incrementNS),
			Running:        running,
			ActiveColor:    parseColor(activeColor),
			LastUpdate:     clockUpdated,
		}
		room.Moves, err = s.loadMoves(ctx, room.ID)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}
	return rooms, rows.Err()
}

func (s *Store) loadMoves(ctx context.Context, roomID string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT san FROM game_moves WHERE room_id = $1 ORDER BY ply`, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var moves []string
	for rows.Next() {
		var move string
		if err := rows.Scan(&move); err != nil {
			return nil, err
		}
		moves = append(moves, move)
	}
	return moves, rows.Err()
}

func emptyToNil(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func colorName(color chess.Color) string {
	switch color {
	case chess.White:
		return "white"
	case chess.Black:
		return "black"
	default:
		return ""
	}
}

func parseColor(value string) chess.Color {
	switch value {
	case "white":
		return chess.White
	case "black":
		return chess.Black
	default:
		return chess.NoColor
	}
}
