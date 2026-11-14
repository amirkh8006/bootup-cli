package tui

import (
	"fmt"
	"strings"

	"github.com/amirkh8006/bootup-cli/internal/prepare"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type preparationModel struct {
	open, review, running, done bool
	cursor                      int
	options                     prepare.Options
	message                     string
	caret                       int
}

type preparationFinished struct{ err error }

func newPreparationModel() preparationModel {
	return preparationModel{open: true, options: prepare.Defaults()}
}

func (p *preparationModel) field() *string {
	switch p.cursor {
	case 0:
		return &p.options.Hostname
	case 1:
		return &p.options.Timezone
	case 2:
		return &p.options.SwapMiB
	case 4:
		return &p.options.SSHPort
	case 5:
		return &p.options.SSHSource
	}
	return nil
}

func (m Model) updatePreparation(msg tea.Msg) (tea.Model, tea.Cmd) {
	p := &m.preparation
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		reserved := 7
		if m.height < 10 {
			reserved = 5
		}
		m.viewportHeight = max(1, m.height-reserved)
	case preparationFinished:
		p.running, p.done = false, true
		p.message = "Preparation completed. Review the log for warnings."
		if msg.err != nil {
			p.message = fmt.Sprintf("Preparation stopped: %v. Earlier steps may have completed.", msg.err)
		}
	case tea.KeyMsg:
		if p.running {
			return m, nil
		}
		key := msg.String()
		if key == "ctrl+c" {
			return m, tea.Quit
		}
		if p.done {
			if key == "enter" || key == "esc" {
				p.open = false
			}
			return m, nil
		}
		if key == "esc" {
			if p.review {
				p.review = false
			} else {
				p.open = false
			}
			p.message = ""
			return m, nil
		}
		if p.review {
			if key == "enter" {
				command, cleanup, err := prepare.Command(p.options)
				if err != nil {
					p.message = err.Error()
					return m, nil
				}
				p.running = true
				return m, tea.ExecProcess(command, func(err error) tea.Msg {
					cleanup()
					return preparationFinished{err: err}
				})
			}
			return m, nil
		}
		switch key {
		case "up", "shift+tab":
			p.cursor = max(0, p.cursor-1)
			if !p.options.Firewall && (p.cursor == 4 || p.cursor == 5) {
				p.cursor = 3
			}
			p.moveCaretToEnd()
		case "down", "tab":
			p.cursor = min(7, p.cursor+1)
			if !p.options.Firewall && (p.cursor == 4 || p.cursor == 5) {
				p.cursor = 6
			}
			p.moveCaretToEnd()
		case "left":
			p.caret = max(0, p.caret-1)
		case "right":
			if field := p.field(); field != nil {
				p.caret = min(len([]rune(*field)), p.caret+1)
			}
		case "home", "ctrl+a":
			p.caret = 0
		case "end", "ctrl+e":
			p.moveCaretToEnd()
		case "enter":
			if p.cursor == 7 {
				if err := p.options.Validate(); err != nil {
					p.message = err.Error()
				} else {
					p.message = ""
					p.review = true
				}
			} else if p.cursor == 3 {
				p.options.Firewall = !p.options.Firewall
			} else if p.cursor == 6 {
				p.options.Upgrade = !p.options.Upgrade
			} else {
				p.cursor++
				p.moveCaretToEnd()
			}
		case " ":
			if p.cursor == 3 {
				p.options.Firewall = !p.options.Firewall
			} else if p.cursor == 6 {
				p.options.Upgrade = !p.options.Upgrade
			}
		case "backspace":
			if field := p.field(); field != nil {
				runes := []rune(*field)
				p.caret = min(p.caret, len(runes))
				if p.caret > 0 {
					*field = string(runes[:p.caret-1]) + string(runes[p.caret:])
					p.caret--
				}
			}
		case "delete":
			if field := p.field(); field != nil {
				runes := []rune(*field)
				if p.caret < len(runes) {
					*field = string(runes[:p.caret]) + string(runes[p.caret+1:])
				}
			}
		case "ctrl+u":
			if field := p.field(); field != nil {
				*field = ""
				p.caret = 0
			}
		default:
			if msg.Type == tea.KeyRunes {
				if field := p.field(); field != nil {
					for _, r := range msg.Runes {
						if r >= 33 && r <= 126 {
							runes := []rune(*field)
							p.caret = min(p.caret, len(runes))
							if len(runes) < 128 {
								*field = string(runes[:p.caret]) + string(r) + string(runes[p.caret:])
								p.caret++
							}
						}
					}
				}
			}
		}
	}
	return m, nil
}

func (p *preparationModel) moveCaretToEnd() {
	if field := p.field(); field != nil {
		p.caret = len([]rune(*field))
	}
}

type preparationField struct {
	group, label, placeholder, help string
}

var preparationFields = []preparationField{
	{"Identity", "Server hostname", "Keep the current hostname", "A short server name, such as production-01. Leave blank to keep the current name."},
	{"Identity", "Timezone", "Keep the current timezone", "Use a timezone such as Asia/Tehran or UTC. Leave blank to keep the current timezone."},
	{"Swap", "Swap size · MiB", "2048", "Create swap only when none exists. 2048 MiB = 2 GiB. Enter 0 to skip creation."},
	{"Firewall", "Configure UFW", "", "Allow SSH, HTTP and HTTPS, with incoming traffic denied by default. Existing rules are preserved."},
	{"Firewall", "SSH port", "22", "The port your SSH server already uses. This adds a firewall rule; it does not change SSH settings."},
	{"Firewall", "SSH source · IP or CIDR", "Allow SSH from any source", "Optional management address, such as 192.0.2.10 or 192.0.2.0/24. Leave blank for any source."},
	{"Packages", "Upgrade installed packages", "", "Common tools and Chrony are always installed. Enable this to also apply available package upgrades."},
	{"Review", "Ready to review", "", "Check all settings on the next screen before starting. You can return here to make changes."},
}

func (m Model) preparationView() string {
	p := m.preparation
	width := 76
	if m.width > 0 {
		width = min(width, m.width-4)
	}
	if width < 26 {
		return "Prepare server\nEnlarge the terminal to at least 30 columns."
	}
	accent := lipgloss.Color("#A78BFA")
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color("#9296A5"))
	label := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#E4E4EF"))
	button := lipgloss.NewStyle().Background(accent).Foreground(lipgloss.Color("#15121E")).Bold(true).Padding(0, 2)
	wrap := lipgloss.NewStyle().Width(width)
	var b strings.Builder
	b.WriteString(titleStyle.Render("Prepare server") + "\n")
	b.WriteString(muted.Render("Server setup  /  Debian & Ubuntu · systemd") + "\n\n")
	if p.done {
		b.WriteString(wrap.Render(p.message) + "\n\n")
		b.WriteString(muted.Render("Log: /var/log/bootup-prepare.log") + "\n\n")
		b.WriteString(button.Render("Return to menu ↵"))
		return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
	}
	value := func(s, fallback string) string {
		if s == "" {
			return fallback
		}
		return s
	}
	onOff := func(v bool) string {
		if v {
			return "Enabled"
		}
		return "Off"
	}
	if p.review {
		b.WriteString(label.Render("Review preparation") + "\n")
		b.WriteString(muted.Render("These settings will be applied to this server.") + "\n\n")
		rows := []struct{ name, value string }{
			{"Hostname", value(p.options.Hostname, "Keep current")},
			{"Timezone", value(p.options.Timezone, "Keep current")},
			{"Swap", p.options.SwapMiB + " MiB · preserve existing swap"},
			{"UFW firewall", onOff(p.options.Firewall)},
			{"Package upgrades", onOff(p.options.Upgrade)},
		}
		if p.options.Firewall {
			rows = append(rows, struct{ name, value string }{"SSH access", "Port " + p.options.SSHPort + " · " + value(p.options.SSHSource, "any source")})
		}
		var summary strings.Builder
		for _, row := range rows {
			summary.WriteString(wrap.Width(width-4).Render(fmt.Sprintf("%-18s %s", row.name, row.value)) + "\n")
		}
		b.WriteString(lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(accent).Padding(0, 1).Width(width - 2).Render(strings.TrimSuffix(summary.String(), "\n")))
		b.WriteString("\n\n")
		if p.options.Firewall {
			b.WriteString(wrap.Width(width).Render("UFW: allow SSH, HTTP & HTTPS; default incoming: deny. Existing rules stay effective.") + "\n")
		}
		b.WriteString(muted.Render("Common tools + Chrony included. Sudo may ask for your password.") + "\n\n")
		b.WriteString(button.Render("Start preparation ↵") + "  " + muted.Render("Esc  Edit settings"))
	} else {
		field := preparationFields[p.cursor]
		groups := []string{"Identity", "Swap", "Firewall", "Packages", "Review"}
		var steps []string
		for _, group := range groups {
			if group == field.group {
				steps = append(steps, lipgloss.NewStyle().Foreground(accent).Bold(true).Render(group))
			} else {
				steps = append(steps, muted.Render(group))
			}
		}
		b.WriteString(wrap.Render(strings.Join(steps, "  ›  ")) + "\n\n")
		b.WriteString(label.Render(field.label) + "\n")
		b.WriteString(wrap.Render(muted.Render(field.help)) + "\n\n")
		if p.cursor == 3 || p.cursor == 6 {
			enabled := p.options.Firewall
			if p.cursor == 6 {
				enabled = p.options.Upgrade
			}
			off, on := "  Off  ", "  Enabled  "
			if enabled {
				on = button.Render("● Enabled")
				off = muted.Render("○ Off")
			} else {
				off = button.Render("● Off")
				on = muted.Render("○ Enabled")
			}
			b.WriteString(off + "    " + on + "\n\n")
			b.WriteString(muted.Render("Space / Enter  Toggle    Tab  Continue"))
		} else if p.cursor == 7 {
			b.WriteString(button.Render("Review settings ↵"))
		} else {
			text := ""
			if f := p.field(); f != nil {
				text = *f
			}
			runes := []rune(text)
			caret := min(p.caret, len(runes))
			// Scroll long values horizontally, keeping the caret in view.
			start := max(0, caret-(width-8))
			end := min(len(runes), start+width-6)
			next := " "
			if caret < len(runes) {
				next = string(runes[caret])
			}
			input := string(runes[start:caret]) + lipgloss.NewStyle().Reverse(true).Render(next)
			if caret < end {
				input += string(runes[caret+1 : end])
			}
			if len(runes) == 0 {
				input += muted.Render(field.placeholder)
			}
			b.WriteString(lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(accent).Padding(0, 1).Width(width-2).Render(input) + "\n")
			b.WriteString(muted.Render("Type to edit · ← → Move cursor · Ctrl+U Clear") + "\n\n")
			b.WriteString(button.Render("Continue ↵"))
		}
		b.WriteString("\n\n" + muted.Render(fmt.Sprintf("%d / 8   Tab / ↓ Next   Shift+Tab / ↑ Back   Esc Menu", p.cursor+1)))
	}
	if p.message != "" {
		b.WriteString("\n\n" + wrap.Width(width).Foreground(lipgloss.Color("#FBBF24")).Render("! "+p.message))
	}
	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}
