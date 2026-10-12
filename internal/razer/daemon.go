package razer

import (
	"fmt"

	"github.com/godbus/dbus/v5"
)

// Through the daemon because it restores these settings at boot.
func daemonDevices() ([]Device, error) {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return nil, err
	}
	var serials []string
	if err := conn.Object("org.razer", "/org/razer").Call("razer.devices.getDevices", 0).Store(&serials); err != nil {
		return nil, fmt.Errorf("openrazer daemon: %w", err)
	}
	var devs []Device
	for _, serial := range serials {
		obj := conn.Object("org.razer", dbus.ObjectPath("/org/razer/device/"+serial))
		var name, kind string
		if err := obj.Call("razer.device.misc.getDeviceName", 0).Store(&name); err != nil {
			continue
		}
		obj.Call("razer.device.misc.getDeviceType", 0).Store(&kind)
		if kind == "headset" {
			continue
		}
		devs = append(devs, Device{name, supported(daemonSettings(obj))})
	}
	return devs, nil
}

func daemonSettings(obj dbus.BusObject) []Setting {
	call := func(method string, args ...any) error { return obj.Call(method, 0, args...).Err }
	maxDPI := int32(16000)
	obj.Call("razer.device.dpi.maxDPI", 0).Store(&maxDPI)
	// The daemon's default for devices that don't list their rates.
	rates := []uint16{125, 500, 1000}
	obj.Call("razer.device.misc.getSupportedPollRates", 0).Store(&rates)
	var pollRates []Option
	for _, r := range rates {
		pollRates = append(pollRates, Option{fmt.Sprintf("%d HZ", r), int(r)})
	}
	settings := []Setting{
		{Section: "SENSOR", Label: "DPI", Min: 100, Max: int(maxDPI), Step: 100,
			Get: func() (int, error) {
				var dpi []int32
				if err := obj.Call("razer.device.dpi.getDPI", 0).Store(&dpi); err != nil || len(dpi) == 0 {
					return Unknown, err
				}
				return int(dpi[0]), nil
			},
			Set: func(v int) error { return call("razer.device.dpi.setDPI", uint16(v), uint16(v)) }},
		{Section: "SENSOR", Label: "POLLING RATE", Get: getInt(obj, "razer.device.misc.getPollRate"),
			Set:     func(v int) error { return call("razer.device.misc.setPollRate", uint16(v)) },
			Options: pollRates},
		{Section: "POWER", Label: "BATTERY", Format: func(v int) string { return fmt.Sprintf("%d%%", v) },
			Get: getFloat(obj, "razer.device.power.getBattery")},
	}
	settings = append(settings, lighting(obj, "LIGHTING", "razer.device.lighting.chroma", "", "razer.device.lighting.brightness")...)
	settings = append(settings, lighting(obj, "LOGO", "razer.device.lighting.logo", "Logo", "razer.device.lighting.logo")...)
	return append(settings, lighting(obj, "SCROLL WHEEL", "razer.device.lighting.scroll", "Scroll", "razer.device.lighting.scroll")...)
}

var effects = []string{"spectrum", "wave", "static"}

func lighting(obj dbus.BusObject, section, iface, zone, brightnessIface string) []Setting {
	method := func(verb, name string) string { return iface + "." + verb + zone + name }
	call := func(method string, args ...any) error { return obj.Call(method, 0, args...).Err }
	return []Setting{
		{Section: section, Label: "BRIGHTNESS", Min: 0, Max: 100, Step: 10,
			Format: func(v int) string { return fmt.Sprintf("%d%%", v) },
			Get:    getFloat(obj, brightnessIface+".get"+zone+"Brightness"),
			Set:    func(v int) error { return call(brightnessIface+".set"+zone+"Brightness", float64(v)) }},
		{Section: section, Label: "EFFECT",
			Options: []Option{{"SPECTRUM", 0}, {"WAVE", 1}, {"STATIC", 2}},
			Get: func() (int, error) {
				var effect string
				if err := obj.Call(method("get", "Effect"), 0).Store(&effect); err != nil {
					return Unknown, err
				}
				for i, e := range effects {
					if e == effect {
						return i, nil
					}
				}
				return Unknown, nil
			},
			Set: func(v int) error {
				switch effects[v] {
				case "spectrum":
					return call(method("set", "Spectrum"))
				case "wave":
					dir := int32(1)
					obj.Call(method("get", "WaveDir"), 0).Store(&dir)
					return call(method("set", "Wave"), dir)
				}
				// shortcut: static reuses the zone's last color, add a color picker when one is wanted
				var colors []byte
				obj.Call(method("get", "EffectColors"), 0).Store(&colors)
				if len(colors) < 3 {
					colors = []byte{0, 255, 0}
				}
				return call(method("set", "Static"), colors[0], colors[1], colors[2])
			}},
	}
}

func getInt(obj dbus.BusObject, method string) func() (int, error) {
	return func() (int, error) {
		var v int32
		if err := obj.Call(method, 0).Store(&v); err != nil {
			return Unknown, err
		}
		return int(v), nil
	}
}

func getFloat(obj dbus.BusObject, method string) func() (int, error) {
	return func() (int, error) {
		var v float64
		if err := obj.Call(method, 0).Store(&v); err != nil {
			return Unknown, err
		}
		return int(v + 0.5), nil
	}
}
