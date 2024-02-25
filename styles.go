package main

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

var (
	colorPrimary     = lipgloss.Color("124")
	colorTextPrimary = lipgloss.Color("254")
	colorTextSubtle  = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	colorSpecial     = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}

	url = lipgloss.NewStyle().
		Foreground(colorSpecial).
		Underline(true).
		Render

	checkMark = lipgloss.NewStyle().SetString("✓").
			Foreground(colorSpecial).
			PaddingRight(1).
			String()

	finished = func(s string) string {
		sty := lipgloss.NewStyle().MarginLeft(0)
		return sty.Render(checkMark + s)
	}

	emph = func(s string) string {
		return lipgloss.NewStyle().Foreground(colorPrimary).Render(s)
	}

	spinStyle = lipgloss.NewStyle().
			MarginLeft(1).
			Foreground(colorPrimary)
)

func decanterFormStyle() *huh.Theme {
	t := huh.ThemeBase()

	f := &t.Focused
	f.Title = lipgloss.NewStyle().Foreground(colorPrimary)
	f.Description = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
	f.Base.Foreground(colorSpecial).BorderForeground(colorPrimary)

	f.Directory = lipgloss.NewStyle().Foreground(colorSpecial)

	f.SelectedOption = lipgloss.NewStyle().Foreground(colorPrimary)
	f.UnselectedOption = lipgloss.NewStyle().Foreground(colorTextPrimary)

	ti := &f.TextInput
	ti.Cursor = lipgloss.NewStyle().Foreground(colorPrimary)

	b := &t.Blurred
	b.MultiSelectSelector = lipgloss.NewStyle().SetString(" ")
	b.SelectSelector = lipgloss.NewStyle().SetString(" ")

	return t
}
