package razer

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const krakenRoot = "/sys/bus/hid/drivers/razerkraken"

// Read directly, since the daemon turns an unreported on/off value into "on".
func headsets(root string) []Device {
	dirs, _ := filepath.Glob(filepath.Join(root, "*:1532:*"))
	var devs []Device
	for _, dir := range dirs {
		if _, err := os.Stat(filepath.Join(dir, "game_chat_balance")); err != nil {
			continue
		}
		name, _ := os.ReadFile(filepath.Join(dir, "device_type"))
		devs = append(devs, Device{strings.TrimSpace(string(name)), supported(headsetSettings(dir))})
	}
	return devs
}

func headsetSettings(dir string) []Setting {
	get := func(name string) func() (int, error) {
		return func() (int, error) { return readInt(filepath.Join(dir, name)) }
	}
	set := func(name string) func(int) error {
		return func(v int) error { return writeInt(filepath.Join(dir, name), v) }
	}
	return []Setting{
		{Section: "SOUND", Label: "EQ PRESET", Get: get("headphone_eq"),
			Set:     func(v int) error { return setEQPreset(dir, v) },
			Options: []Option{{"DEFAULT", 0}, {"GAME", 1}, {"MOVIE", 2}, {"MUSIC", 3}, {"ESPORTS", 4}}},
		{Section: "SOUND", Label: "THX SPATIAL AUDIO", Options: onOff,
			Get: get("thx_spatial_audio"), Set: set("thx_spatial_audio")},
		{Section: "SOUND", Label: "GAME / CHAT BALANCE", Min: 0, Max: 20, Step: 1, Format: balance,
			Get: get("game_chat_balance"), Set: set("game_chat_balance")},
		{Section: "SOUND", Label: "BLUETOOTH CALL MIX", Get: get("in_call_audio_mix"), Set: set("in_call_audio_mix"),
			Options: []Option{{"COMBINE", 0}, {"LOWER 2.4 GHZ", 1}, {"MUTE 2.4 GHZ", 2}}},

		{Section: "MIC", Label: "SIDETONE", Min: 0, Max: 15, Step: 1,
			Get: get("sidetone"), Set: set("sidetone")},
		{Section: "MIC", Label: "MIC EQ", Get: get("mic_eq_preset"), Set: set("mic_eq_preset"),
			Options: []Option{{"DEFAULT", 0}, {"ESPORTS", 1}, {"BROADCAST", 2}, {"MIC BOOST", 3}}},
		{Section: "MIC", Label: "MIC", Get: get("mic_mute"),
			Options: []Option{{"LIVE", 0}, {"MUTED", 1}}},

		{Section: "HEADSET", Label: "BATTERY", Format: func(v int) string { return fmt.Sprintf("%d%%", v) },
			Get: func() (int, error) { return battery(filepath.Join(dir, "charge_level")) }},
		{Section: "HEADSET", Label: "CHARGING", Get: get("charge_status"),
			Options: []Option{{"NO", 0}, {"YES", 1}}},
		{Section: "HEADSET", Label: "ULTRA-LOW LATENCY", Options: onOff,
			Get: get("ultra_low_latency"), Set: set("ultra_low_latency")},
		{Section: "HEADSET", Label: "AUTO POWER OFF", Get: get("wireless_power_save"), Set: set("wireless_power_save"),
			Options: []Option{{"15 MIN", 15}, {"30 MIN", 30}, {"45 MIN", 45}, {"60 MIN", 60}}},
		{Section: "HEADSET", Label: "AUDIO PROMPTS", Options: onOff,
			Get: get("audio_prompts"), Set: set("audio_prompts")},
		{Section: "HEADSET", Label: "FUNCTION BUTTON", Get: get("audio_function_button"), Set: set("audio_function_button"),
			Options: []Option{{"SIDETONE", 1}, {"FOOTSTEPS", 2}}},
	}
}

// Writing a preset also stores its bands, so a wrong read-back would overwrite it.
func setEQPreset(dir string, preset int) error {
	slot := filepath.Join(dir, "eq_slot")
	if err := writeInt(slot, preset); err != nil {
		return err
	}
	stored, err := os.ReadFile(slot)
	if err != nil {
		return err
	}
	if f := strings.Fields(string(stored)); len(f) != 11 || f[0] != strconv.Itoa(preset) {
		return fmt.Errorf("headset didn't return the bands of preset %d", preset)
	}
	return os.WriteFile(filepath.Join(dir, "headphone_eq"), stored, 0o644)
}

func battery(path string) (int, error) {
	v, err := readInt(path)
	if err != nil || v == Unknown {
		return Unknown, err
	}
	return (v*100 + 127) / 255, nil
}

func balance(v int) string {
	switch {
	case v < 10:
		return fmt.Sprintf("GAME +%d", 10-v)
	case v > 10:
		return fmt.Sprintf("CHAT +%d", v-10)
	}
	return "EVEN"
}

func readInt(path string) (int, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Unknown, err
	}
	f := strings.Fields(string(b))
	if len(f) == 0 {
		return Unknown, fmt.Errorf("%s is empty", path)
	}
	return strconv.Atoi(f[0])
}

func writeInt(path string, v int) error {
	return os.WriteFile(path, []byte(strconv.Itoa(v)), 0o644)
}
