package razer

import (
	"slices"
	"testing"

	"github.com/godbus/dbus/v5"
)

// fakeObj answers D-Bus calls from replies and records the rest.
type fakeObj struct {
	dbus.BusObject
	replies map[string][]any
	calls   []string
	args    [][]any
}

func (f *fakeObj) Call(method string, _ dbus.Flags, args ...any) *dbus.Call {
	f.calls, f.args = append(f.calls, method), append(f.args, args)
	return &dbus.Call{Body: f.replies[method]}
}

func TestStaticColor(t *testing.T) {
	obj := &fakeObj{replies: map[string][]any{
		"razer.device.lighting.logo.getLogoEffectColors": {[]byte{0, 255, 0, 0, 0, 0}},
	}}
	var color Setting
	for _, s := range lighting(obj, "LOGO", "razer.device.lighting.logo", "Logo", "razer.device.lighting.logo") {
		if s.Label == "STATIC COLOR" {
			color = s
		}
	}
	if v, err := color.Get(); v != 0x00FF00 || err != nil {
		t.Fatalf("got %06X %v, want the reported green", v, err)
	}
	if err := color.Set(0x8000FF); err != nil {
		t.Fatal(err)
	}
	last := len(obj.calls) - 1
	if obj.calls[last] != "razer.device.lighting.logo.setLogoStatic" || !slices.Equal(obj.args[last], []any{byte(0x80), byte(0), byte(0xFF)}) {
		t.Errorf("set sent %s %v, want setLogoStatic 128 0 255", obj.calls[last], obj.args[last])
	}
}

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
