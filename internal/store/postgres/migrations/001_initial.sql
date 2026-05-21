CREATE TABLE IF NOT EXISTS schema_migrations (
	version text PRIMARY KEY,
	applied_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS players (
	id uuid PRIMARY KEY DEFAULT uuidv7(),
	nickname text NOT NULL,
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS player_keys (
	fingerprint text PRIMARY KEY,
	player_id uuid NOT NULL REFERENCES players(id) ON DELETE CASCADE,
	authorized_key text NOT NULL,
	key_type text NOT NULL,
	created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS games (
	room_id text PRIMARY KEY,
	status text NOT NULL,
	base_ns bigint NOT NULL,
	increment_ns bigint NOT NULL,
	white_player_id uuid REFERENCES players(id),
	white_name text NOT NULL DEFAULT '',
	black_player_id uuid REFERENCES players(id),
	black_name text NOT NULL DEFAULT '',
	turn text NOT NULL DEFAULT 'white',
	fen text NOT NULL DEFAULT '',
	white_remaining_ns bigint NOT NULL,
	black_remaining_ns bigint NOT NULL,
	clock_running boolean NOT NULL DEFAULT false,
	active_color text NOT NULL DEFAULT '',
	clock_updated_at timestamptz NOT NULL DEFAULT now(),
	outcome text NOT NULL DEFAULT '',
	method text NOT NULL DEFAULT '',
	last_event text NOT NULL DEFAULT '',
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now(),
	finished_at timestamptz
);

CREATE TABLE IF NOT EXISTS game_moves (
	room_id text NOT NULL REFERENCES games(room_id) ON DELETE CASCADE,
	ply integer NOT NULL,
	player_id uuid REFERENCES players(id),
	san text NOT NULL,
	from_square text NOT NULL,
	to_square text NOT NULL,
	fen_after text NOT NULL,
	white_remaining_ns bigint NOT NULL,
	black_remaining_ns bigint NOT NULL,
	played_at timestamptz NOT NULL DEFAULT now(),
	PRIMARY KEY (room_id, ply)
);

CREATE TABLE IF NOT EXISTS game_events (
	id bigserial PRIMARY KEY,
	room_id text REFERENCES games(room_id) ON DELETE CASCADE,
	event_type text NOT NULL,
	player_id uuid REFERENCES players(id),
	message text NOT NULL DEFAULT '',
	created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS chat_messages (
	id bigserial PRIMARY KEY,
	room_id text REFERENCES games(room_id) ON DELETE SET NULL,
	player_id uuid REFERENCES players(id) ON DELETE SET NULL,
	body text NOT NULL,
	created_at timestamptz NOT NULL DEFAULT now(),
	deleted_at timestamptz
);

CREATE TABLE IF NOT EXISTS ratings (
	player_id uuid PRIMARY KEY REFERENCES players(id) ON DELETE CASCADE,
	rating integer NOT NULL DEFAULT 1200,
	deviation integer NOT NULL DEFAULT 350,
	updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS rating_events (
	id bigserial PRIMARY KEY,
	player_id uuid NOT NULL REFERENCES players(id) ON DELETE CASCADE,
	room_id text REFERENCES games(room_id) ON DELETE SET NULL,
	before_rating integer NOT NULL,
	after_rating integer NOT NULL,
	reason text NOT NULL,
	created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS bots (
	id uuid PRIMARY KEY DEFAULT uuidv7(),
	player_id uuid NOT NULL REFERENCES players(id) ON DELETE CASCADE,
	name text NOT NULL,
	config jsonb NOT NULL DEFAULT '{}'::jsonb,
	enabled boolean NOT NULL DEFAULT true,
	created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS tournaments (
	id uuid PRIMARY KEY DEFAULT uuidv7(),
	name text NOT NULL,
	status text NOT NULL DEFAULT 'draft',
	config jsonb NOT NULL DEFAULT '{}'::jsonb,
	starts_at timestamptz,
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS tournament_entries (
	tournament_id uuid NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
	player_id uuid NOT NULL REFERENCES players(id) ON DELETE CASCADE,
	seed integer,
	score numeric(6, 2) NOT NULL DEFAULT 0,
	created_at timestamptz NOT NULL DEFAULT now(),
	PRIMARY KEY (tournament_id, player_id)
);

CREATE TABLE IF NOT EXISTS tournament_games (
	tournament_id uuid NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
	room_id text NOT NULL REFERENCES games(room_id) ON DELETE CASCADE,
	round integer NOT NULL,
	board integer NOT NULL,
	created_at timestamptz NOT NULL DEFAULT now(),
	PRIMARY KEY (tournament_id, room_id)
);

CREATE INDEX IF NOT EXISTS games_status_idx ON games(status);
CREATE INDEX IF NOT EXISTS game_moves_room_ply_idx ON game_moves(room_id, ply);
CREATE INDEX IF NOT EXISTS game_events_room_created_idx ON game_events(room_id, created_at);
