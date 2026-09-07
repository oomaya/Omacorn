package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Sharp-Corner Aesthetic Law compliant styling
var (
	// Palettes
	ColorBg     = lipgloss.Color("#0b0f19")
	ColorCard   = lipgloss.Color("#161e2e")
	ColorCyan   = lipgloss.Color("#00f0ff")
	ColorGreen  = lipgloss.Color("#2ed573")
	ColorRed    = lipgloss.Color("#ff4757")
	ColorOrange = lipgloss.Color("#ffa502")
	ColorPurple = lipgloss.Color("#a55eea")
	ColorMuted  = lipgloss.Color("#747d8c")
	ColorWhite  = lipgloss.Color("#f1f2f6")

	// Sharp-edge border
	SharpBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "┌",
		TopRight:    "┐",
		BottomLeft:  "└",
		BottomRight: "┘",
	}

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorCyan).
			Padding(0, 1)

	BadgeActive = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorWhite).
			Background(ColorGreen).
			Padding(0, 1)

	BadgeInactive = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorWhite).
			Background(ColorMuted).
			Padding(0, 1)

	BadgeWarning = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorWhite).
			Background(ColorOrange).
			Padding(0, 1)

	CardStyle = lipgloss.NewStyle().
			Border(SharpBorder).
			BorderForeground(lipgloss.Color("#2f3542")).
			Padding(0, 1).
			Margin(0, 0, 1, 0)

	ActiveCardStyle = lipgloss.NewStyle().
			Border(SharpBorder).
			BorderForeground(ColorCyan).
			Padding(0, 1).
			Margin(0, 0, 1, 0)

	TabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorMuted).
			Padding(0, 2)

	ActiveTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorCyan).
			Background(lipgloss.Color("#1e272e")).
			Padding(0, 2)

	HelpKeyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorCyan)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)
)
