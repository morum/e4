package clock

import (
	"time"

	"github.com/morum/e4/internal/domain"

	"github.com/notnil/chess"
)

type State struct {
	whiteRemaining time.Duration
	blackRemaining time.Duration
	increment      time.Duration
	running        bool
	activeColor    chess.Color
	lastUpdate     time.Time
}

func New(tc domain.TimeControl) State {
	return State{
		whiteRemaining: tc.Base,
		blackRemaining: tc.Base,
		increment:      tc.Increment,
	}
}

func (s *State) Start(turn chess.Color, now time.Time) {
	s.running = true
	s.activeColor = turn
	s.lastUpdate = now
}

func (s *State) Stop(now time.Time) {
	s.applyElapsed(now)
	s.running = false
	s.lastUpdate = now
}

func (s *State) Switch(turn chess.Color, now time.Time) {
	s.applyElapsed(now)
	s.addIncrement(turn)
	s.running = true
	s.activeColor = turn.Other()
	s.lastUpdate = now
}

func (s *State) Snapshot(now time.Time) (time.Duration, time.Duration) {
	white := s.whiteRemaining
	black := s.blackRemaining
	if !s.running {
		return clampDuration(white), clampDuration(black)
	}

	elapsed := now.Sub(s.lastUpdate)
	if elapsed < 0 {
		elapsed = 0
	}

	if s.activeColor == chess.White {
		white -= elapsed
	} else if s.activeColor == chess.Black {
		black -= elapsed
	}

	return clampDuration(white), clampDuration(black)
}

func (s *State) Flagged(now time.Time) (chess.Color, bool) {
	if !s.running {
		return chess.NoColor, false
	}

	white, black := s.Snapshot(now)
	if white <= 0 {
		return chess.White, true
	}
	if black <= 0 {
		return chess.Black, true
	}
	return chess.NoColor, false
}

func (s *State) applyElapsed(now time.Time) {
	if !s.running {
		return
	}

	elapsed := now.Sub(s.lastUpdate)
	if elapsed < 0 {
		elapsed = 0
	}

	if s.activeColor == chess.White {
		s.whiteRemaining -= elapsed
	} else if s.activeColor == chess.Black {
		s.blackRemaining -= elapsed
	}

	s.whiteRemaining = clampDuration(s.whiteRemaining)
	s.blackRemaining = clampDuration(s.blackRemaining)
	s.lastUpdate = now
}

func (s *State) addIncrement(color chess.Color) {
	if color == chess.White {
		s.whiteRemaining += s.increment
	} else if color == chess.Black {
		s.blackRemaining += s.increment
	}
}

func clampDuration(d time.Duration) time.Duration {
	if d < 0 {
		return 0
	}
	return d
}
