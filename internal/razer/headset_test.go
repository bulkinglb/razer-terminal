package razer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHeadsets(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "0003:1532:057A.0001")
	files := map[string]string{
		"device_type":       "Razer BlackShark V3 (Wireless)\n",
		"game_chat_balance": "-1\n",
		"headphone_eq":      "-1 0 0 0 0 0 0 0 0 0 0\n",
		"charge_level":      "247\n",
		"sidetone":          "4\n",
	}
	os.Mkdir(dir, 0o755)
	os.Mkdir(filepath.Join(root, "0003:1532:0504.0002"), 0o755) // a Kraken without V3 files
	for name, content := range files {
		os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644)
	}

	devs := headsets(root)
	if len(devs) != 1 || devs[0].Name != "Razer BlackShark V3 (Wireless)" {
		t.Fatalf("got %+v, want just the V3", devs)
	}
	got := map[string]int{}
	for _, s := range devs[0].Settings {
		got[s.Label], _ = s.Get()
	}
	want := map[string]int{"EQ PRESET": Unknown, "GAME / CHAT BALANCE": Unknown, "SIDETONE": 4, "BATTERY": 97}
	if len(got) != len(want) {
		t.Errorf("settings %v, want only those with a file: %v", got, want)
	}
	for label, v := range want {
		if got[label] != v {
			t.Errorf("%s = %d, want %d", label, got[label], v)
		}
	}
}

func TestSetEQPresetChecksSlot(t *testing.T) {
	dir := t.TempDir()
	eq := filepath.Join(dir, "headphone_eq")
	os.WriteFile(eq, []byte("1 2 2 5 5 1 -1 2 3 3 3\n"), 0o644)
	// A plain file echoes the slot without bands, like a missing reply.
	if err := setEQPreset(dir, 2); err == nil {
		t.Fatal("want an error when the bands don't come back")
	}
	if b, _ := os.ReadFile(eq); string(b) != "1 2 2 5 5 1 -1 2 3 3 3\n" {
		t.Errorf("headphone_eq changed to %q", b)
	}
}

func TestLowLatencyRemembersWrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ultra_low_latency")
	os.WriteFile(path, []byte("1\n"), 0o644)
	var ull Setting
	for _, s := range headsetSettings(filepath.Dir(path)) {
		if s.Label == "ULTRA-LOW LATENCY" {
			ull = s
		}
	}
	os.WriteFile(path, []byte("0\n"), 0o644) // later reads must not hit the file
	if v, err := ull.Get(); v != 1 || err != nil {
		t.Fatalf("before a write: %d, %v; want the startup value 1", v, err)
	}
	if err := ull.Set(0); err != nil {
		t.Fatal(err)
	}
	if v, _ := ull.Get(); v != 0 {
		t.Errorf("after writing 0: %d", v)
	}
	if b, _ := os.ReadFile(path); string(b) != "0" {
		t.Errorf("file holds %q, want 0", b)
	}
}
