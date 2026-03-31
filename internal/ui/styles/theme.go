package styles

import (
	"charm.land/bubbles/v2/table"
	"charm.land/lipgloss/v2"
)

type Theme struct {
	Doc     lipgloss.Style
	Title   lipgloss.Style
	Help    lipgloss.Style
	Search  lipgloss.Style
	Preview lipgloss.Style
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

		Preview: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1),
	}
}

type TableTheme struct {
	HeaderBorderColor string
	SelectedFg        string
	SelectedBg        string
}

func DefaultTableTheme() TableTheme {
	return TableTheme{
		HeaderBorderColor: "240",
		SelectedFg:        "229",
		SelectedBg:        "57",
	}
}

// TableStyles converts a TableTheme into a bubbles table.Styles value,
// performing lipgloss.Color conversions in one place.
func TableStyles(t TableTheme) table.Styles {
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(t.HeaderBorderColor)).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color(t.SelectedFg)).
		Background(lipgloss.Color(t.SelectedBg)).
		Bold(false)
	return s
}
