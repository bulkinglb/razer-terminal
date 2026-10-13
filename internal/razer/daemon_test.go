package razer

import (
	"slices"
	"testing"
)

func TestEditStage(t *testing.T) {
	two := []dpiStage{{800, 800}, {1600, 1600}}
	tests := []struct {
		name       string
		stages     []dpiStage
		i, dpi     int
		wantActive byte
		want       []dpiStage
		wantErr    bool
	}{
		{"change", two, 0, 900, 1, []dpiStage{{900, 900}, {1600, 1600}}, false},
		{"add next", two, 2, 2400, 3, []dpiStage{{800, 800}, {1600, 1600}, {2400, 2400}}, false},
		{"skip a stage", two, 3, 2400, 2, nil, true},
		{"remove", two, 0, 0, 1, []dpiStage{{1600, 1600}}, false},
		{"already off", two, 4, 0, 2, nil, false},
		{"remove last one", two[:1], 0, 0, 2, nil, true},
	}
	for _, tt := range tests {
		active, got, err := editStage(2, tt.stages, tt.i, tt.dpi)
		if active != tt.wantActive || !slices.Equal(got, tt.want) || (err != nil) != tt.wantErr {
			t.Errorf("%s: got %d %v %v, want %d %v err=%v", tt.name, active, got, err, tt.wantActive, tt.want, tt.wantErr)
		}
	}
	if two[0].X != 800 {
		t.Errorf("editStage changed its input: %v", two)
	}
}
