package tui

import (
	"strings"
	"testing"

	"github.com/amirkh8006/bootup-cli/internal/prepare"
	tea "github.com/charmbracelet/bubbletea"
)

func TestPreparationNavigation(t *testing.T) {
	m := Model{height: 30, width: 100, services: []Service{{Name: "Prepare server"}, {Name: "nginx"}}}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	if !m.preparation.open || m.quitting {
		t.Fatal("preparation should open inside the TUI")
	}
	m.preparation.cursor = 7
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	if !m.preparation.review {
		t.Fatal("expected review before execution")
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)
	if m.preparation.review || !m.preparation.open {
		t.Fatal("escape should return to settings")
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)
	if m.preparation.open || m.quitting {
		t.Fatal("escape should return to menu")
	}
}

func TestPreparationAcceptsLettersUsedAsMenuShortcuts(t *testing.T) {
	m := Model{preparation: newPreparationModel()}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("hqjk")})
	m = updated.(Model)
	if m.preparation.options.Hostname != "hqjk" || m.quitting || m.showHelp {
		t.Fatal("typing should edit the form")
	}
}

func TestPreparationFailureReturnsToResultScreen(t *testing.T) {
	m := Model{preparation: preparationModel{open: true, running: true, options: prepare.Defaults()}}
	updated, _ := m.Update(preparationFinished{err: nil})
	m = updated.(Model)
	if !m.preparation.done || m.preparation.running || !strings.Contains(m.View(), "bootup-prepare.log") {
		t.Fatal("expected preparation result and log path")
	}
}

func TestPreparationEditsAtCursor(t *testing.T) {
	m := Model{preparation: newPreparationModel()}
	for _, key := range []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune("prod-01")},
		{Type: tea.KeyHome},
		{Type: tea.KeyRunes, Runes: []rune("eu-")},
		{Type: tea.KeyDelete},
	} {
		updated, _ := m.Update(key)
		m = updated.(Model)
	}
	if got := m.preparation.options.Hostname; got != "eu-rod-01" {
		t.Fatalf("editing at caret produced %q", got)
	}
}

func TestPreparationSkipsInactiveFirewallFields(t *testing.T) {
	m := Model{preparation: newPreparationModel()}
	m.preparation.cursor = 3
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(Model)
	if m.preparation.cursor != 6 {
		t.Fatal("SSH fields should be skipped when firewall changes are off")
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m = updated.(Model)
	if m.preparation.cursor != 3 {
		t.Fatal("back navigation should return to firewall toggle")
	}
}
