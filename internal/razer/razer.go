package razer

const Unknown = -1

type Option struct {
	Label string
	Value int
}

type Setting struct {
	Section        string
	Label          string
	Min, Max, Step int
	Options        []Option
	Format         func(int) string
	Get            func() (int, error)
	Set            func(int) error // nil means read-only
}

type Device struct {
	Name     string
	Settings []Setting
}

// Returns the headsets even when the daemon fails.
func Devices() ([]Device, error) {
	devs := headsets(krakenRoot)
	more, err := daemonDevices()
	return append(devs, more...), err
}

func supported(settings []Setting) []Setting {
	var out []Setting
	for _, s := range settings {
		if _, err := s.Get(); err == nil {
			out = append(out, s)
		}
	}
	return out
}

var onOff = []Option{{"OFF", 0}, {"ON", 1}}
