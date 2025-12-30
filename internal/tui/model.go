package tui

import (
	"github.com/amirkh8006/bootup-cli/internal/services"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Service represents a bootup service
type Service struct {
	Name        string
	Description string
	Category    string
	Installed   bool
	Installing  bool
}

// Model represents the TUI state
type Model struct {
	services        []Service
	cursor          int
	width           int
	height          int
	viewportTop     int // Top line of the viewport
	viewportHeight  int // Number of lines visible in viewport
	installing      bool
	installMsg      string
	quitting        bool
	selectedService string // Service selected for installation
}

// InstallationMsg represents an installation status message
type InstallationMsg struct {
	Service string
	Success bool
	Error   error
}

// NewModel creates a new TUI model
func NewModel() Model {
	var tuiServices []Service

	// Get services organized by category to maintain consistent ordering
	servicesByCategory := services.GetServicesByCategory()
	categoryOrder := services.GetCategoryOrder()

	// Add services in the same order as they're displayed in the view
	for _, category := range categoryOrder {
		categoryServices, exists := servicesByCategory[category]
		if !exists {
			continue
		}

		for _, serviceInfo := range categoryServices {
			tuiServices = append(tuiServices, Service{
				Name:        serviceInfo.Name,
				Description: serviceInfo.Description,
				Category:    serviceInfo.Category,
				Installed:   services.IsServiceInstalled(serviceInfo.Name),
				Installing:  false,
			})
		}
	}

	return Model{
		services: tuiServices,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#04B575"))

	categoryStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFA500")).
			Underline(true)

	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#EE6FF8"))

	selectedInstalledStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#04B575")).
				Padding(0, 1)

	installedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575"))

	installedServiceStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6B8E67")).
				Italic(true)

	installingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFAA00"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))
)

// adjustViewport adjusts the viewport to ensure the cursor is visible
func (m Model) adjustViewport() Model {
	if m.viewportHeight <= 0 || len(m.services) == 0 {
		return m
	}

	// Calculate which line the cursor is on in the rendered output
	cursorLine := m.getCursorLinePosition()

	// Calculate the line where the current service's category header starts
	categoryHeaderLine := m.getCategoryHeaderLine(m.cursor)

	// Adjust viewport if cursor goes out of bounds
	if cursorLine < m.viewportTop {
		// When scrolling up, ensure we show the category header if possible
		desiredTop := categoryHeaderLine
		// But make sure the cursor is still visible
		if cursorLine-desiredTop >= m.viewportHeight {
			desiredTop = cursorLine - m.viewportHeight + 1
		}
		m.viewportTop = desiredTop
	} else if cursorLine >= m.viewportTop+m.viewportHeight {
		m.viewportTop = cursorLine - m.viewportHeight + 1
	}

	// Ensure viewport doesn't go negative or beyond content
	m.viewportTop = max(0, m.viewportTop)

	return m
}

// getCategoryHeaderLine returns the line position of the category header for the given service index
func (m Model) getCategoryHeaderLine(serviceIndex int) int {
	if serviceIndex < 0 || serviceIndex >= len(m.services) {
		return 0
	}

	targetCategory := m.services[serviceIndex].Category
	lineIndex := 0
	currentCategory := ""

	for i, service := range m.services {
		// Add lines for category headers
		if service.Category != currentCategory {
			if currentCategory != "" {
				lineIndex++ // Empty line between categories
			}

			// This is where the category header line starts
			if service.Category == targetCategory {
				return lineIndex
			}

			lineIndex++ // Category header line
			currentCategory = service.Category
		}

		// If we've found our service, we've gone too far
		if i == serviceIndex {
			break
		}
		lineIndex++
	}

	return 0
}

// getCursorLinePosition calculates which line the cursor is on in the rendered output
func (m Model) getCursorLinePosition() int {
	lineIndex := 0
	currentCategory := ""

	for i, service := range m.services {
		// Add lines for category headers
		if service.Category != currentCategory {
			if currentCategory != "" {
				lineIndex++ // Empty line between categories
			}
			lineIndex++ // Category header line
			currentCategory = service.Category
		}

		// This is the line for the current service
		if i == m.cursor {
			return lineIndex
		}
		lineIndex++
	}

	return lineIndex
}
