package tui

import (
	"fmt"
	"strings"

	"github.com/amirkh8006/bootup-cli/internal/services"
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

	// Organize services by category
	servicesByCategory := make(map[string][]Service)
	for _, service := range m.services {
		servicesByCategory[service.Category] = append(servicesByCategory[service.Category], service)
	}

	// Display services grouped by category
	categoryOrder := services.GetCategoryOrder()
	currentIndex := 0

	for _, category := range categoryOrder {
		categoryServices, exists := servicesByCategory[category]
		if !exists || len(categoryServices) == 0 {
			continue
		}

		// Category header
		b.WriteString(categoryStyle.Render(category + ":"))
		b.WriteString("\n")

		// Services in this category
		for _, service := range categoryServices {
			cursor := "  "
			if m.cursor == currentIndex {
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

			if m.cursor == currentIndex {
				line = selectedStyle.Render(line)
			}

			b.WriteString(line)
			b.WriteString("\n")
			currentIndex++
		}
		b.WriteString("\n")
	}

	// Instructions
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("Controls: ↑/↓ or j/k: navigate • space/enter: install • q: quit"))

	return b.String()
}
