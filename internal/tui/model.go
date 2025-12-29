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
	for _, serviceInfo := range services.GetAllServices() {
		tuiServices = append(tuiServices, Service{
			Name:        serviceInfo.Name,
			Description: serviceInfo.Description,
			Category:    serviceInfo.Category,
			Installed:   false,
			Installing:  false,
		})
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

	installedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575"))

	installingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFAA00"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))
)
