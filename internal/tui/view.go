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

	// Show help overlay if requested
	if m.showHelp {
		help := `⌨️  Keyboard Shortcuts

Navigation:
  ↑, k         Move up
  ↓, j         Move down
  PgUp, Ctrl+B Page up
  PgDn, Ctrl+F Page down
  Home, g      Go to first service
  End, G       Go to last service
  1-9          Quick jump to service (1st-9th)

Actions:
  Space, Enter Install selected service
  ?, h         Show/hide this help
  q, Esc, ^C   Quit

Press any key to close help...`

		return lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(1, 2).
			Margin(1, 2).
			Render(help)
	}

	// Handle extremely small terminals
	if m.height < 5 {
		return fmt.Sprintf("🚀 Bootup\nTerminal too small (%dx%d)\nMinimum: 5 lines", m.width, m.height)
	}

	var b strings.Builder

	// For very small terminals, use a compact header
	if m.height < 10 {
		b.WriteString(titleStyle.Render("🚀 Bootup CLI"))
		b.WriteString("\n")
	} else {
		// Header
		b.WriteString(titleStyle.Render("🚀 Bootup CLI - Interactive Service Installer"))
		b.WriteString("\n\n")
	}

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

	// Compact header for small terminals
	if m.height < 10 {
		b.WriteString(headerStyle.Render("Services:"))
		b.WriteString("\n")
	} else {
		b.WriteString(headerStyle.Render("Available Services:"))
		b.WriteString("\n\n")
	}

	// Build the service list with categories
	var serviceLines []string
	currentCategory := ""

	for i, service := range m.services {
		// Add category header when we encounter a new category
		if service.Category != currentCategory {
			if currentCategory != "" {
				serviceLines = append(serviceLines, "") // Add spacing between categories
			}
			serviceLines = append(serviceLines, categoryStyle.Render(service.Category+":"))
			currentCategory = service.Category
		}

		cursor := "  "
		if m.cursor == i {
			cursor = "▶ " // Arrow for selected services (both installed and not)
		}

		status := ""
		name := service.Name

		if service.Installing {
			status = installingStyle.Render(" (installing...)")
		} else if service.Installed {
			status = installedStyle.Render(" (installed ✓)")
		}

		line := fmt.Sprintf("%s  %s - %s%s",
			cursor, name, service.Description, status)

		if m.cursor == i {
			line = selectedStyle.Render(line)
		}

		serviceLines = append(serviceLines, line)
	}

	// Calculate viewport bounds
	viewportEnd := min(m.viewportTop+m.viewportHeight, len(serviceLines))
	viewportStart := max(0, m.viewportTop)

	// Render only the visible portion
	visibleLines := serviceLines[viewportStart:viewportEnd]
	for _, line := range visibleLines {
		b.WriteString(line)
		b.WriteString("\n")
	}

	// Add scroll indicators if needed
	if len(serviceLines) > m.viewportHeight {
		scrollInfo := fmt.Sprintf("(Showing %d-%d of %d services)",
			viewportStart+1, viewportEnd, len(serviceLines))
		b.WriteString("\n")
		b.WriteString(helpStyle.Render(scrollInfo))
	}

	// Instructions - compact for small terminals
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("Press ? or h for help"))

	return b.String()
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
