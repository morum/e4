package render

import "github.com/morum/e4/internal/domain"

type StatusKind string

const (
	StatusInfo    StatusKind = "info"
	StatusSuccess StatusKind = "success"
	StatusWarning StatusKind = "warning"
	StatusError   StatusKind = "error"
)

type StatusLine struct {
	Kind    StatusKind
	Message string
}

type Orientation string

const (
	OrientationWhite Orientation = "white"
	OrientationBlack Orientation = "black"
)

type LayoutMode string

const (
	LayoutNarrow  LayoutMode = "narrow"
	LayoutCompact LayoutMode = "compact"
	LayoutWide    LayoutMode = "wide"
)

type Context struct {
	Width       int
	Height      int
	ANSI        bool
	Role        domain.Role
	Orientation Orientation
	Status      StatusLine
}

func (c Context) Layout() LayoutMode {
	width := c.Width
	if width <= 0 {
		width = 80
	}

	switch {
	case width >= 110:
		return LayoutWide
	case width >= 84:
		return LayoutCompact
	default:
		return LayoutNarrow
	}
}

func OrientationForRole(role domain.Role) Orientation {
	if role == domain.RoleBlack {
		return OrientationBlack
	}
	return OrientationWhite
}
