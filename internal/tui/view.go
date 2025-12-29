package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the TUI
func (m Model) View() string {
	if m.quitting {
		return "Thanks for using Bootup CLI! 👋\n"
	}

	var b strings.Builder

	// Header
	b.WriteString(titleStyle.Render("🚀 Bootup CLI - Interactive Service Installer"))
	b.WriteString("\n\n")

	if m.installing {
		b.WriteString(installingStyle.Render(fmt.Sprintf("Preparing to install %s...", m.selectedService)))
		b.WriteString("\nExiting TUI to perform installation in normal terminal mode.\n")
		return b.String()
	}

	if m.installMsg != "" {
		if strings.Contains(m.installMsg, "✓") {
			b.WriteString(installedStyle.Render(m.installMsg))
		} else {
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Render(m.installMsg))
		}
		b.WriteString("\n\n")
	}

	b.WriteString(headerStyle.Render("Available Services:"))
	b.WriteString("\n\n")

	// Display services in the same order as stored in m.services
	currentCategory := ""

	for i, service := range m.services {
		// Show category header when we encounter a new category
		if service.Category != currentCategory {
			if currentCategory != "" {
				b.WriteString("\n") // Add spacing between categories
			}
			b.WriteString(categoryStyle.Render(service.Category + ":"))
			b.WriteString("\n")
			currentCategory = service.Category
		}

		cursor := "  "
		if m.cursor == i {
			cursor = "▶ "
		}

		status := ""
		name := service.Name

		if service.Installing {
			status = installingStyle.Render(" (installing...)")
		} else if service.Installed {
			status = installedStyle.Render(" ✓")
		}

		line := fmt.Sprintf("%s  %s - %s%s",
			cursor, name, service.Description, status)

		if m.cursor == i {
			line = selectedStyle.Render(line)
		}

		b.WriteString(line)
		b.WriteString("\n")
	}

	// Instructions
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("Controls: ↑/↓ or j/k: navigate • space/enter: install • q: quit"))

	return b.String()
}
