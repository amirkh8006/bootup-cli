package services

import (
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/charmbracelet/lipgloss"
)

// ── styles ────────────────────────────────────────────────────────────────────

var (
	titleBoxStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 2)

	sectionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			MarginTop(1)

	dividerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3D3D3D"))

	okStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#04B575"))

	warnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFAA00"))

	failStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#DDDDDD")).
			Width(16)

	portOkStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575"))

	portFailStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555"))
)

// ── data types ────────────────────────────────────────────────────────────────

type CheckResult struct {
	Name    string
	Status  string // "ok", "warn", "fail", "info"
	Message string
	Pct     float64 // 0 means no bar
}

type ServiceHealth struct {
	ServiceName string
	Installed   bool
	Running     bool
	Ports       []PortCheck
}

type PortCheck struct {
	Port    int
	Purpose string
	Open    bool
}

// ── service metadata ──────────────────────────────────────────────────────────

var serviceSystemdName = map[string]string{
	"nginx":             "nginx",
	"caddy":             "caddy",
	"postgresql":        "postgresql",
	"mongodb":           "mongod",
	"redis":             "redis-server",
	"elasticsearch":     "elasticsearch",
	"mysql":             "mysql",
	"clickhouse":        "clickhouse-server",
	"kafka":             "kafka",
	"rabbitmq":          "rabbitmq-server",
	"prometheus":        "prometheus",
	"grafana":           "grafana-server",
	"alertmanager":      "alertmanager",
	"docker":            "docker",
	"rustfs":            "rustfs",
	"seaweedfs":         "weed",
	"mongodb_exporter":  "mongodb_exporter",
	"nginx_exporter":    "nginx_exporter",
	"node_exporter":     "node_exporter",
	"postgres_exporter": "postgres_exporter",
	"redis_exporter":    "redis_exporter",
}

var servicePorts = map[string][]PortCheck{
	"nginx":             {{80, "HTTP", false}, {443, "HTTPS", false}},
	"caddy":             {{80, "HTTP", false}, {443, "HTTPS", false}},
	"postgresql":        {{5432, "DB", false}},
	"mongodb":           {{27017, "DB", false}},
	"redis":             {{6379, "DB", false}},
	"elasticsearch":     {{9200, "HTTP", false}, {9300, "Transport", false}},
	"mysql":             {{3306, "DB", false}},
	"clickhouse":        {{8123, "HTTP", false}, {9000, "Native", false}},
	"kafka":             {{9092, "Broker", false}},
	"rabbitmq":          {{5672, "AMQP", false}, {15672, "Mgmt", false}},
	"prometheus":        {{9090, "HTTP", false}},
	"grafana":           {{3000, "HTTP", false}},
	"alertmanager":      {{9093, "HTTP", false}},
	"docker":            {{2376, "API", false}},
	"rustfs":            {{9000, "S3", false}},
	"seaweedfs":         {{9333, "Master", false}},
	"mongodb_exporter":  {{9216, "Metrics", false}},
	"nginx_exporter":    {{9113, "Metrics", false}},
	"node_exporter":     {{9100, "Metrics", false}},
	"postgres_exporter": {{9187, "Metrics", false}},
	"redis_exporter":    {{9121, "Metrics", false}},
}

// ── entry point ───────────────────────────────────────────────────────────────

func RunDoctor() {
	fmt.Println()
	fmt.Println(titleBoxStyle.Render("  🩺  Bootup Doctor  "))
	fmt.Println()

	renderSystemSection()
	renderServicesSection()
	renderFailedUnitsSection()

	fmt.Println()
}

// ── system checks ─────────────────────────────────────────────────────────────

func renderSystemSection() {
	fmt.Println(sectionStyle.Render("  System Health"))
	fmt.Println(dividerStyle.Render("  " + strings.Repeat("─", 54)))

	disk := checkDiskSpace()
	mem := checkMemory()

	renderSystemRow(disk)
	renderSystemRow(mem)
}

func renderSystemRow(r CheckResult) {
	icon := statusDot(r.Status)
	label := labelStyle.Render(r.Name)
	msg := dimStyle.Render(r.Message)

	bar := ""
	if r.Pct > 0 {
		bar = "  " + renderBar(r.Pct, 20, r.Status)
	}

	fmt.Printf("  %s  %s%s  %s\n", icon, label, bar, msg)
}

func renderBar(pct float64, width int, status string) string {
	filled := int((pct / 100.0) * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled

	filledStr := strings.Repeat("█", filled)
	emptyStr := strings.Repeat("░", empty)

	var filledStyled string
	switch status {
	case "fail":
		filledStyled = failStyle.Render(filledStr)
	case "warn":
		filledStyled = warnStyle.Render(filledStr)
	default:
		filledStyled = okStyle.Render(filledStr)
	}

	return filledStyled + dimStyle.Render(emptyStr)
}

func checkDiskSpace() CheckResult {
	var stat syscall.Statfs_t
	if err := syscall.Statfs("/", &stat); err != nil {
		return CheckResult{"Disk Space", "warn", "unable to check", 0}
	}

	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bfree * uint64(stat.Bsize)
	usedPct := 100 - (float64(free)/float64(total))*100

	msg := fmt.Sprintf("%.0f%% used · %.1f GB free", usedPct, float64(free)/1e9)

	status := "ok"
	if usedPct >= 95 {
		status = "fail"
	} else if usedPct >= 85 {
		status = "warn"
	}
	return CheckResult{"Disk Space", status, msg, usedPct}
}

func checkMemory() CheckResult {
	out, err := exec.Command("free", "-m").Output()
	if err != nil {
		return CheckResult{"Memory", "info", "unavailable (non-Linux?)", 0}
	}

	for _, line := range strings.Split(string(out), "\n") {
		if !strings.HasPrefix(line, "Mem:") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 7 {
			break
		}
		total, _ := strconv.ParseInt(fields[1], 10, 64)
		available, _ := strconv.ParseInt(fields[6], 10, 64)
		usedPct := 100 - (float64(available)/float64(total))*100
		msg := fmt.Sprintf("%.0f%% used · %d MB free", usedPct, available)

		status := "ok"
		if usedPct >= 95 {
			status = "fail"
		} else if usedPct >= 85 {
			status = "warn"
		}
		return CheckResult{"Memory", status, msg, usedPct}
	}
	return CheckResult{"Memory", "info", "unable to parse", 0}
}

// ── service checks ────────────────────────────────────────────────────────────

func renderServicesSection() {
	fmt.Println()
	fmt.Println(sectionStyle.Render("  Installed Services"))
	fmt.Println(dividerStyle.Render("  " + strings.Repeat("─", 54)))

	var found []ServiceHealth

	categoryOrder := GetCategoryOrder()
	byCategory := GetServicesByCategory()

	for _, cat := range categoryOrder {
		svcs, ok := byCategory[cat]
		if !ok {
			continue
		}
		for _, svc := range svcs {
			if !IsServiceInstalled(svc.Name) {
				continue
			}
			h := ServiceHealth{
				ServiceName: svc.Name,
				Installed:   true,
				Running:     isServiceActiveByName(svc.Name),
			}
			if portDefs, ok := servicePorts[svc.Name]; ok {
				for _, pd := range portDefs {
					pd.Open = isPortListening(pd.Port)
					h.Ports = append(h.Ports, pd)
				}
			}
			found = append(found, h)
		}
	}

	if len(found) == 0 {
		fmt.Println(dimStyle.Render("  No bootup-managed services found."))
		return
	}

	for _, svc := range found {
		renderServiceRow(svc)
	}
}

func renderServiceRow(svc ServiceHealth) {
	var statusStr string
	var dot string

	if svc.Running {
		dot = okStyle.Render("●")
		statusStr = okStyle.Render("running ")
	} else {
		dot = failStyle.Render("○")
		statusStr = failStyle.Render("stopped")
	}

	name := lipgloss.NewStyle().Width(20).Render(svc.ServiceName)
	ports := renderPorts(svc.Ports)

	fmt.Printf("  %s  %s  %s  %s\n", dot, name, statusStr, ports)
}

func renderPorts(ports []PortCheck) string {
	if len(ports) == 0 {
		return ""
	}
	var parts []string
	for _, p := range ports {
		portStr := fmt.Sprintf(":%d", p.Port)
		if p.Open {
			parts = append(parts, portOkStyle.Render(portStr+"✓"))
		} else {
			parts = append(parts, portFailStyle.Render(portStr+"✗"))
		}
	}
	return dimStyle.Render("[") + strings.Join(parts, dimStyle.Render(" ")) + dimStyle.Render("]")
}

// ── failed units ──────────────────────────────────────────────────────────────

func renderFailedUnitsSection() {
	units := getFailedSystemdUnits()
	if len(units) == 0 {
		return
	}

	fmt.Println()
	fmt.Println(warnStyle.Bold(true).Render("  ⚠  Failed Systemd Units"))
	fmt.Println(dividerStyle.Render("  " + strings.Repeat("─", 54)))
	for _, u := range units {
		fmt.Printf("  %s  %s\n", failStyle.Render("✗"), failStyle.Render(u))
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func isServiceActiveByName(serviceName string) bool {
	systemdName, ok := serviceSystemdName[serviceName]
	if !ok {
		return false
	}
	if checkSystemdActive(systemdName) {
		return true
	}
	// fallback versioned names for postgresql / redis
	if serviceName == "postgresql" {
		for _, v := range []string{"17", "16", "15", "14"} {
			if checkSystemdActive("postgresql-" + v) {
				return true
			}
		}
	}
	if serviceName == "redis" {
		return checkSystemdActive("redis")
	}
	return false
}

func checkSystemdActive(name string) bool {
	out, err := exec.Command("systemctl", "is-active", name).Output()
	return err == nil && strings.TrimSpace(string(out)) == "active"
}

func isPortListening(port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 300*1000*1000)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func getFailedSystemdUnits() []string {
	out, err := exec.Command("systemctl", "--failed", "--no-legend", "--no-pager").Output()
	if err != nil {
		return nil
	}
	var units []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) > 0 {
			units = append(units, fields[0])
		}
	}
	return units
}

func statusDot(status string) string {
	switch status {
	case "ok":
		return okStyle.Render("✓")
	case "warn":
		return warnStyle.Render("⚠")
	case "fail":
		return failStyle.Render("✗")
	default:
		return dimStyle.Render("·")
	}
}
