package domain

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type TimeControl struct {
	Base      time.Duration
	Increment time.Duration
}

func ParseTimeControl(input string) (TimeControl, error) {
	parts := strings.Split(strings.TrimSpace(input), "|")
	if len(parts) != 2 {
		return TimeControl{}, fmt.Errorf("invalid time control %q: expected <minutes>|<increment>", input)
	}

	minutes, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil || minutes <= 0 {
		return TimeControl{}, fmt.Errorf("invalid base minutes in %q", input)
	}

	increment, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil || increment < 0 {
		return TimeControl{}, fmt.Errorf("invalid increment seconds in %q", input)
	}

	return TimeControl{
		Base:      time.Duration(minutes) * time.Minute,
		Increment: time.Duration(increment) * time.Second,
	}, nil
}

func (tc TimeControl) String() string {
	return fmt.Sprintf("%d|%d", int(tc.Base/time.Minute), int(tc.Increment/time.Second))
}
