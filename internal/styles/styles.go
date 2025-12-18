package styles

import "github.com/charmbracelet/lipgloss"

const (
	SidebarWidth = 22
	MaxOutput    = 1000 // Max lines per terminal
)

var (
	Cyan    = lipgloss.Color("14")
	Green   = lipgloss.Color("10")
	Red     = lipgloss.Color("9")
	Yellow  = lipgloss.Color("11")
	Magenta = lipgloss.Color("13")
	Gray    = lipgloss.Color("8")
	Blue    = lipgloss.Color("12")
	White   = lipgloss.Color("15")
	BgDark  = lipgloss.Color("235")

	Title    = lipgloss.NewStyle().Foreground(Cyan).Bold(true)
	Prompt   = lipgloss.NewStyle().Foreground(Green).Bold(true)
	Profile  = lipgloss.NewStyle().Foreground(Magenta)
	Path     = lipgloss.NewStyle().Foreground(Blue)
	Git      = lipgloss.NewStyle().Foreground(Yellow)
	Error    = lipgloss.NewStyle().Foreground(Red)
	Success  = lipgloss.NewStyle().Foreground(Green)
	Muted    = lipgloss.NewStyle().Foreground(Gray)
	Selected = lipgloss.NewStyle().Foreground(Green).Bold(true)
	Normal   = lipgloss.NewStyle().Foreground(White)
	Running  = lipgloss.NewStyle().Foreground(Yellow).Bold(true)

	Sidebar = lipgloss.NewStyle().
		Width(SidebarWidth).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(Gray).
		Padding(0, 1)

	Pane = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(Cyan)

	StatusBar = lipgloss.NewStyle().
			Background(BgDark).
			Foreground(White)
)
