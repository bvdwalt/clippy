package styles

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Doc    lipgloss.Style
	Title  lipgloss.Style
	Help   lipgloss.Style
	Search lipgloss.Style
}

func DefaultTheme() Theme {
	return Theme{
		Doc: lipgloss.NewStyle().Margin(1, 2),

		Title: lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true).
			Padding(0, 1),

		Help: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Margin(1, 0),

		Search: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1).
			Width(50),
	}
}

type TableTheme struct {
	HeaderBorderColor lipgloss.Color
	SelectedFg        lipgloss.Color
	SelectedBg        lipgloss.Color
}

func DefaultTableTheme() TableTheme {
	return TableTheme{
		HeaderBorderColor: lipgloss.Color("240"),
		SelectedFg:        lipgloss.Color("229"),
		SelectedBg:        lipgloss.Color("57"),
	}
}
