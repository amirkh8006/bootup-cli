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

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quitting = true
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.services)-1 {
				m.cursor++
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
