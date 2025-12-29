package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// Run starts the TUI application
func Run() error {
	model := NewModel()
	p := tea.NewProgram(model, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	// Check if we need to perform any installation after TUI exits
	if m, ok := finalModel.(Model); ok && m.selectedService != "" {
		fmt.Printf("\n🚀 Installing %s...\n", m.selectedService)

		// Get the appropriate installer
		installer := GetServiceInstaller(m.selectedService)

		// Perform installation in normal terminal mode
		if err := installer(); err != nil {
			fmt.Printf("❌ Failed to install %s: %v\n", m.selectedService, err)
			return err
		}

		fmt.Printf("✅ %s installed successfully!\n", m.selectedService)
	}

	return nil
}
