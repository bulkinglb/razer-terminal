package tui

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/bulkinglb/razer-terminal/internal/razer"
)

const labelWidth, valueWidth = 26, 16

// Reread so changes made on the headset itself show up.
const refreshEvery = 2 * time.Second

type model struct {
	devs   []razer.Device
	values [][]int
	tab    int
	row    int
	status string
}

type tickMsg struct{}

func Run(devs []razer.Device, status string) error {
	_, err := tea.NewProgram(newModel(devs, status)).Run()
	return err
}

func newModel(devs []razer.Device, status string) model {
	m := model{devs: devs, values: make([][]int, len(devs)), status: status}
	for i, d := range devs {
		m.values[i] = make([]int, len(d.Settings))
		m.refresh(i)
	}
	return m
}

func (m model) refresh(dev int) {
	for i, s := range m.devs[dev].Settings {
		v, err := s.Get()
		if err != nil {
			v = razer.Unknown
		}
		m.values[dev][i] = v
	}
}

func tick() tea.Cmd {
	return tea.Tick(refreshEvery, func(time.Time) tea.Msg { return tickMsg{} })
}

func (m model) Init() tea.Cmd { return tick() }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		last := len(m.devs[m.tab].Settings) - 1
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			m = m.switchTab(m.tab + 1)
		case "shift+tab":
			m = m.switchTab(m.tab - 1)
		case "up", "k":
			m.row = max(0, m.row-1)
		case "down", "j":
			m.row = max(0, min(last, m.row+1))
		case "left", "h":
			return m.adjust(-1, false), nil
		case "right", "l":
			return m.adjust(1, false), nil
		case "shift+left", "H":
			return m.adjust(-1, true), nil
		case "shift+right", "L":
			return m.adjust(1, true), nil
		case "r":
			m.refresh(m.tab)
		}
	case tickMsg:
		m.refresh(m.tab)
		return m, tick()
	}
	return m, nil
}

func (m model) switchTab(t int) model {
	n := len(m.devs)
	m.tab, m.row, m.status = (t%n+n)%n, 0, ""
	m.refresh(m.tab)
	return m
}

func (m model) adjust(dir int, fine bool) model {
	settings := m.devs[m.tab].Settings
	if m.row >= len(settings) || settings[m.row].Set == nil {
		return m
	}
	s := settings[m.row]
	v := next(s, m.values[m.tab][m.row], dir, fine)
	if err := s.Set(v); err != nil {
		m.status = "WRITE FAILED: " + err.Error()
		return m
	}
	m.values[m.tab][m.row] = v
	m.status = ""
	return m
}

// Unreported values start mid-range so the first press can't jump to an extreme.
func next(s razer.Setting, v, dir int, fine bool) int {
	if len(s.Options) > 0 {
		i := slices.IndexFunc(s.Options, func(o razer.Option) bool { return o.Value == v })
		if i < 0 {
			return s.Options[0].Value
		}
		return s.Options[(i+dir+len(s.Options))%len(s.Options)].Value
	}
	if v == razer.Unknown {
		return (s.Min + s.Max) / 2
	}
	step := s.Step
	if fine {
		step = max(1, step/10)
	}
	return min(max(v+dir*step, s.Min), s.Max)
}

func text(s razer.Setting, v int) string {
	if v == razer.Unknown {
		return "UNKNOWN"
	}
	for _, o := range s.Options {
		if o.Value == v {
			return o.Label
		}
	}
	if s.Format != nil {
		return s.Format(v)
	}
	return strconv.Itoa(v)
}

func detail(s razer.Setting, v int) string {
	switch {
	case v == razer.Unknown && s.Set != nil:
		return "NOT REPORTED YET · CHANGING IT SETS IT"
	case v == razer.Unknown:
		return "NOT REPORTED YET"
	case s.Set == nil:
		return "READ ONLY"
	case len(s.Options) == 0:
		return fmt.Sprintf("%d (%d-%d)", v, s.Min, s.Max)
	}
	return ""
}

func (m model) View() tea.View {
	var b strings.Builder
	b.WriteString("\x1b[1m◆ RAZER TERMINAL\x1b[22m\n")
	for i, d := range m.devs {
		if i > 0 {
			b.WriteString(" ")
		}
		name := " " + strings.ToUpper(d.Name) + " "
		if i == m.tab {
			name = "\x1b[7m" + name + "\x1b[27m"
		}
		b.WriteString(name)
	}
	b.WriteString("\n\n")

	settings, values := m.devs[m.tab].Settings, m.values[m.tab]
	if len(settings) == 0 {
		b.WriteString("  NOTHING TO SET\n")
	}
	for i, s := range settings {
		if i == 0 || s.Section != settings[i-1].Section {
			if i > 0 {
				b.WriteString("\n")
			}
			b.WriteString("\x1b[1m" + s.Section + "\x1b[22m\n")
		}
		label := s.Label + strings.Repeat(".", max(2, labelWidth-len([]rune(s.Label))))
		if i == m.row {
			label = "\x1b[7m" + label + "\x1b[27m"
		}
		left, right := "<", ">"
		if s.Set == nil {
			left, right = " ", " "
		}
		fmt.Fprintf(&b, "  %s %s %s %s\n", label, left, center(text(s, values[i]), valueWidth), right)
	}

	status := m.status
	if status == "" && m.row < len(settings) {
		status = "> " + settings[m.row].Label + "  " + detail(settings[m.row], values[m.row])
	}
	b.WriteString("\n" + status + "\n")
	b.WriteString("\x1b[2m↑↓ SELECT · ←→ CHANGE · SHIFT+←→ FINE · TAB DEVICE · R REFRESH · Q QUIT\x1b[22m")

	v := tea.NewView(b.String())
	v.AltScreen = true
	return v
}

func center(s string, width int) string {
	r := []rune(s)
	r = r[:min(len(r), width)]
	pad := width - len(r)
	return strings.Repeat(" ", pad/2) + string(r) + strings.Repeat(" ", pad-pad/2)
}
