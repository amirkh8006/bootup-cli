package iran

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type iranScreen int

const (
	screenCategory iranScreen = iota
	screenMirror
)

// IranModel is the bubbletea model for the iran TUI
type IranModel struct {
	screen       iranScreen
	catCursor    int
	mirrorCursor int
	width        int
	height       int
	quitting     bool

	SelectedCat    int // -1 = none
	SelectedMirror int // -1 = none
}

var (
	iranTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	iranCatHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFA500"))

	iranSelectedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#EE6FF8"))

	iranActiveStatusStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#04B575"))

	iranDefaultStatusStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#626262"))

	iranURLStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#5C9BF5"))

	iranDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	iranHelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))
)

func NewIranModel() IranModel {
	return IranModel{
		screen:         screenCategory,
		SelectedCat:    -1,
		SelectedMirror: -1,
	}
}

func (m IranModel) Init() tea.Cmd { return nil }

func (m IranModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch m.screen {
		case screenCategory:
			switch msg.String() {
			case "ctrl+c", "q", "esc":
				m.quitting = true
				return m, tea.Quit
			case "up", "k":
				if m.catCursor > 0 {
					m.catCursor--
				}
			case "down", "j":
				if m.catCursor < len(Categories)-1 {
					m.catCursor++
				}
			case "enter", " ":
				m.screen = screenMirror
				m.mirrorCursor = 0
			}

		case screenMirror:
			cat := Categories[m.catCursor]
			switch msg.String() {
			case "ctrl+c", "q":
				m.quitting = true
				return m, tea.Quit
			case "esc", "b":
				m.screen = screenCategory
			case "up", "k":
				if m.mirrorCursor > 0 {
					m.mirrorCursor--
				}
			case "down", "j":
				if m.mirrorCursor < len(cat.Mirrors)-1 {
					m.mirrorCursor++
				}
			case "enter", " ":
				m.SelectedCat = m.catCursor
				m.SelectedMirror = m.mirrorCursor
				m.quitting = true
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m IranModel) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	b.WriteString(iranTitleStyle.Render("Iran Mirror Manager"))
	b.WriteString("\n\n")

	switch m.screen {
	case screenCategory:
		b.WriteString(iranCatHeaderStyle.Render("Select a category:"))
		b.WriteString("\n\n")

		for i, cat := range Categories {
			cursor := "  "
			if m.catCursor == i {
				cursor = "▶ "
			}

			status := cat.Status()
			var statusStr string
			if status != "default" && status != "unknown" && status != "npm not installed" && status != "go not installed" {
				statusStr = iranActiveStatusStyle.Render(" [" + status + "]")
			} else {
				statusStr = iranDefaultStatusStyle.Render(" [" + status + "]")
			}

			namePart := cat.Name
			if m.catCursor == i {
				namePart = iranSelectedStyle.Render(namePart)
			}

			b.WriteString(fmt.Sprintf("%s  %s%s\n", cursor, namePart, statusStr))
		}

		b.WriteString("\n")
		b.WriteString(iranHelpStyle.Render("↑↓/jk navigate   Enter select   q quit"))

	case screenMirror:
		cat := Categories[m.catCursor]

		b.WriteString(iranCatHeaderStyle.Render(cat.Name))
		b.WriteString("\n\n")

		for i, mirror := range cat.Mirrors {
			cursor := "  "
			if m.mirrorCursor == i {
				cursor = "▶ "
			}

			name := fmt.Sprintf("%-14s", mirror.Name)
			url := iranURLStyle.Render(mirror.URL)
			desc := iranDescStyle.Render("  — " + mirror.Description)

			if m.mirrorCursor == i {
				name = iranSelectedStyle.Render(name)
			}

			b.WriteString(fmt.Sprintf("%s  %s  %s%s\n", cursor, name, url, desc))
		}

		b.WriteString("\n")
		b.WriteString(iranHelpStyle.Render("↑↓/jk navigate   Enter select   Esc/b back   q quit"))
	}

	return b.String()
}

// RunTUI launches the iran mirror TUI and applies the selected mirror after exit
func RunTUI() error {
	p := tea.NewProgram(NewIranModel(), tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	m, ok := finalModel.(IranModel)
	if !ok || m.SelectedCat < 0 || m.SelectedMirror < 0 {
		return nil
	}

	cat := Categories[m.SelectedCat]
	mirror := cat.Mirrors[m.SelectedMirror]

	fmt.Printf("\nSetting %s mirror to %s...\n", cat.Name, mirror.Name)

	var opErr error
	if mirror.IsDefault {
		opErr = cat.Reset()
	} else {
		opErr = cat.Apply(mirror)
	}

	if opErr != nil {
		fmt.Printf("❌ Failed: %v\n", opErr)
		return opErr
	}

	fmt.Printf("✅ %s mirror set to %s\n", cat.Key, mirror.Name)
	return nil
}
