package tui

import (
	"testing"

	"github.com/bulkinglb/razer-terminal/internal/razer"
)

func TestNext(t *testing.T) {
	balance := razer.Setting{Min: 0, Max: 20, Step: 1}
	dpi := razer.Setting{Min: 100, Max: 26000, Step: 100}
	power := razer.Setting{Options: []razer.Option{{Label: "15 MIN", Value: 15}, {Label: "30 MIN", Value: 30}, {Label: "60 MIN", Value: 60}}}
	tests := []struct {
		name         string
		s            razer.Setting
		v, dir, want int
		fine         bool
	}{
		{"unknown range starts in the middle", balance, razer.Unknown, -1, 10, false},
		{"range stops at its end", balance, 20, 1, 20, false},
		{"coarse step", dpi, 1800, 1, 1900, false},
		{"fine step", dpi, 1800, -1, 1790, true},
		{"choices wrap forward", power, 60, 1, 15, false},
		{"choices wrap back", power, 15, -1, 60, false},
		{"unknown choice starts at the first", power, razer.Unknown, -1, 15, false},
	}
	for _, tt := range tests {
		if got := next(tt.s, tt.v, tt.dir, tt.fine); got != tt.want {
			t.Errorf("%s: got %d, want %d", tt.name, got, tt.want)
		}
	}
}

func TestLinkedRereads(t *testing.T) {
	// Editing a DPI stage makes it the active one, so the DPI row changes too.
	dpi := 800
	get := func() (int, error) { return dpi, nil }
	stage := razer.Setting{Min: 0, Max: 2000, Step: 100, Linked: true, Get: get,
		Set: func(v int) error { dpi = v; return nil }}
	plain := stage
	plain.Linked = false
	shown := razer.Setting{Get: get}

	for _, tt := range []struct {
		s    razer.Setting
		want int
	}{{stage, 900}, {plain, 800}} {
		dpi = 800
		m := newModel([]razer.Device{{Name: "mouse", Settings: []razer.Setting{tt.s, shown}}}, "")
		m = m.adjust(1, false)
		if m.values[0][0] != 900 || m.values[0][1] != tt.want {
			t.Errorf("linked=%v: rows show %v, want [900 %d]", tt.s.Linked, m.values[0], tt.want)
		}
	}
}
