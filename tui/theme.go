package tui

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Name   string
	Bg     lipgloss.Color
	Bg2    lipgloss.Color
	Bg3    lipgloss.Color
	Fg     lipgloss.Color
	Muted  lipgloss.Color
	Faint  lipgloss.Color
	Red    lipgloss.Color
	Orange lipgloss.Color
	Green  lipgloss.Color
	Blue   lipgloss.Color
	Purple lipgloss.Color
}

type Styles struct {
	MainContainer  lipgloss.Style
	Header         lipgloss.Style
	Label          lipgloss.Style
	Path           lipgloss.Style
	Help           lipgloss.Style
	Key            lipgloss.Style
	Footer         lipgloss.Style
	TableContainer lipgloss.Style
	Error          lipgloss.Style
	Success        lipgloss.Style
	TableHeader    lipgloss.Style
	TableSelected  lipgloss.Style
	TableCell      lipgloss.Style
}

var FlexokiDark = Theme{
	Name:   "Flexoki Dark",
	Bg:     lipgloss.Color("#100F0F"),
	Bg2:    lipgloss.Color("#1C1B1A"),
	Bg3:    lipgloss.Color("#282726"),
	Fg:     lipgloss.Color("#CECDC3"),
	Muted:  lipgloss.Color("#878580"),
	Faint:  lipgloss.Color("#575653"),
	Red:    lipgloss.Color("#D14D41"),
	Orange: lipgloss.Color("#DA702C"),
	Green:  lipgloss.Color("#879A39"),
	Blue:   lipgloss.Color("#4385BE"),
	Purple: lipgloss.Color("#8B7EC8"),
}

var FlexokiLight = Theme{
	Name:   "Flexoki Light",
	Bg:     lipgloss.Color("#FFFCF0"),
	Bg2:    lipgloss.Color("#F2F0E5"),
	Bg3:    lipgloss.Color("#E6E4D9"),
	Fg:     lipgloss.Color("#100F0F"),
	Muted:  lipgloss.Color("#6F6E69"),
	Faint:  lipgloss.Color("#B7B5AC"),
	Red:    lipgloss.Color("#AF3029"),
	Orange: lipgloss.Color("#BC5215"),
	Green:  lipgloss.Color("#66800B"),
	Blue:   lipgloss.Color("#205EA6"),
	Purple: lipgloss.Color("#5E409D"),
}

func InitStyles(t Theme) Styles {
	s := Styles{}

	s.MainContainer = lipgloss.NewStyle().
		Padding(1, 2).
		Foreground(t.Fg)

	s.Header = lipgloss.NewStyle().
		Bold(true).
		Foreground(t.Bg).
		Background(t.Red).
		Padding(0, 1).
		MarginBottom(1)

	s.Label = lipgloss.NewStyle().
		Foreground(t.Fg).
		Bold(true)

	s.Path = lipgloss.NewStyle().
		Foreground(t.Fg).
		Italic(true).
		Underline(true)

	s.Help = lipgloss.NewStyle().
		Foreground(t.Muted)

	s.Key = lipgloss.NewStyle().
		Foreground(t.Orange).
		Bold(true)

	s.Footer = lipgloss.NewStyle().
		MarginTop(1).
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(t.Bg2).
		PaddingTop(1)

	s.TableContainer = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(t.Bg2).
		MarginTop(1)

	s.Error = lipgloss.NewStyle().
		Foreground(t.Red).
		Bold(true)

	s.Success = lipgloss.NewStyle().
		Foreground(t.Green).
		Bold(true)

	// Table Styles
	s.TableHeader = lipgloss.NewStyle().
		Foreground(t.Fg).
		Bold(true).
		Padding(0, 1).
		Border(lipgloss.NormalBorder(), false, true, true, false).
		BorderForeground(t.Bg2)

	s.TableSelected = lipgloss.NewStyle().
		Foreground(t.Fg).
		Background(t.Bg3)

	s.TableCell = lipgloss.NewStyle().
		Foreground(t.Fg).
		Padding(0, 1).
		Border(lipgloss.NormalBorder(), false, true, false, false).
		BorderForeground(t.Bg2)

	return s
}
