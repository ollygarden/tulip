package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	Green  = lipgloss.Color("#00cc66")
	Yellow = lipgloss.Color("#ffcc00")
	Red    = lipgloss.Color("#ff3333")
	Blue   = lipgloss.Color("#3399ff")
	Gray   = lipgloss.Color("#888888")
	White  = lipgloss.Color("#ffffff")

	// Styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(White)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(Gray)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(Green)

	WarningStyle = lipgloss.NewStyle().
			Foreground(Yellow)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(Red)

	InfoStyle = lipgloss.NewStyle().
			Foreground(Blue)

	DimStyle = lipgloss.NewStyle().
			Foreground(Gray)

	BoldStyle = lipgloss.NewStyle().
			Bold(true)

	// Table styles
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(White).
			Padding(0, 1)

	CellStyle = lipgloss.NewStyle().
			Padding(0, 1)

	// Status indicators
	CheckMark = SuccessStyle.Render("✓")
	CrossMark = ErrorStyle.Render("✗")
	Warning   = WarningStyle.Render("⚠")
	Info      = InfoStyle.Render("ℹ")
	Bullet    = DimStyle.Render("●")
)

// FormatSuccess formats a success message.
func FormatSuccess(msg string) string {
	return fmt.Sprintf("  %s %s", CheckMark, msg)
}

// FormatError formats an error message.
func FormatError(msg string) string {
	return fmt.Sprintf("  %s %s", CrossMark, msg)
}

// FormatWarning formats a warning message.
func FormatWarning(msg string) string {
	return fmt.Sprintf("  %s %s", Warning, msg)
}

// FormatInfo formats an info message.
func FormatInfo(msg string) string {
	return fmt.Sprintf("  %s %s", Info, msg)
}

// FormatHeader formats a section header with a line underneath.
func FormatHeader(title string) string {
	return fmt.Sprintf("\n  ── %s ──────────────────────────────────────────", TitleStyle.Render(title))
}

// StabilityColor returns the appropriate color for a stability level.
func StabilityColor(stability string) lipgloss.Color {
	switch stability {
	case "stable":
		return Green
	case "beta":
		return Yellow
	case "alpha":
		return lipgloss.Color("#ff9933")
	case "development":
		return Red
	case "deprecated":
		return Gray
	default:
		return Gray
	}
}

// FormatStability returns a colored stability label.
func FormatStability(stability string) string {
	color := StabilityColor(stability)
	return lipgloss.NewStyle().Foreground(color).Render(stability)
}

// FormatSource returns a styled source label (base vs custom).
func FormatSource(source string) string {
	switch source {
	case "base":
		return InfoStyle.Render("base")
	case "custom":
		return WarningStyle.Render("custom")
	default:
		return DimStyle.Render(source)
	}
}
