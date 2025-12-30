package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// Update handles TUI state updates
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Calculate viewport height based on terminal size:
		// For small terminals (height < 10): reserve 5 lines for UI
		// For normal terminals: reserve 7 lines for UI
		reservedLines := 7
		if m.height < 10 {
			reservedLines = 5
		}
		m.viewportHeight = max(1, m.height-reservedLines)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quitting = true
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				// Special case: if we're now at the first service, show from top
				if m.cursor == 0 {
					m.viewportTop = 0
				} else {
					m = m.adjustViewport()
				}
			}

		case "down", "j":
			if m.cursor < len(m.services)-1 {
				m.cursor++
				m = m.adjustViewport()
			}

		case "pgup", "ctrl+b":
			// Page up - move cursor up by viewport height
			if len(m.services) > 0 {
				newCursor := max(0, m.cursor-m.viewportHeight)
				m.cursor = newCursor
				m = m.adjustViewport()
			}

		case "pgdown", "ctrl+f":
			// Page down - move cursor down by viewport height
			if len(m.services) > 0 {
				newCursor := min(len(m.services)-1, m.cursor+m.viewportHeight)
				m.cursor = newCursor
				m = m.adjustViewport()
			}

		case "home", "g":
			// Go to first service
			if len(m.services) > 0 {
				m.cursor = 0
				m.viewportTop = 0 // Always show from the top when going to first service
			}

		case "end", "G":
			// Go to last service
			if len(m.services) > 0 {
				m.cursor = len(m.services) - 1
				m = m.adjustViewport()
			}

		case " ", "enter":
			if !m.installing {
				// Set installing state and quit TUI to perform installation in normal terminal
				m.installing = true
				m.installMsg = fmt.Sprintf("Installing %s...", m.services[m.cursor].Name)

				// Store the selected service for installation after TUI exit
				m.selectedService = m.services[m.cursor].Name
				m.quitting = true
				return m, tea.Quit
			}
		}

	case InstallationMsg:
		// This is no longer used since we handle installation after TUI exit
		m.installing = false
	}

	return m, nil
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
