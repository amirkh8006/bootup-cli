package iran

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/amirkh8006/bootup-cli/internal/utils"
)

type Mirror struct {
	Name        string
	URL         string
	Description string
	IsDefault   bool
	Meta        map[string]string // extra per-mirror config (e.g. "security_url" for apt)
}

type MirrorCategory struct {
	Name    string
	Key     string
	Mirrors []Mirror
	Apply   func(Mirror) error
	Reset   func() error
	Status  func() string
}

var Categories []MirrorCategory

func init() {
	Categories = []MirrorCategory{
		{
			Name:    "APT (Ubuntu/Debian packages)",
			Key:     "apt",
			Mirrors: aptMirrors,
			Apply:   applyAPT,
			Reset:   resetAPT,
			Status:  statusAPT,
		},
		{
			Name:    "npm (Node.js packages)",
			Key:     "npm",
			Mirrors: npmMirrors,
			Apply:   applyNPM,
			Reset:   resetNPM,
			Status:  statusNPM,
		},
		{
			Name:    "pip (Python packages)",
			Key:     "pip",
			Mirrors: pipMirrors,
			Apply:   applyPIP,
			Reset:   resetPIP,
			Status:  statusPIP,
		},
		{
			Name:    "Docker registry",
			Key:     "docker",
			Mirrors: dockerMirrors,
			Apply:   applyDocker,
			Reset:   resetDocker,
			Status:  statusDocker,
		},
		{
			Name:    "Go module proxy",
			Key:     "golang",
			Mirrors: goMirrors,
			Apply:   applyGo,
			Reset:   resetGo,
			Status:  statusGo,
		},
	}
}

var aptMirrors = []Mirror{
	{
		Name:        "Default",
		URL:         "http://archive.ubuntu.com/ubuntu",
		Description: "Ubuntu official (restore original sources.list)",
		IsDefault:   true,
	},
	{
		Name:        "Liara",
		URL:         "http://linux-mirror.liara.ir/repository/ubuntu",
		Description: "Liara cloud mirror",
		Meta:        map[string]string{"security_url": "http://linux-mirror.liara.ir/repository/ubuntu-security"},
	},
	{
		Name:        "ArvanCloud",
		URL:         "http://mirror.arvancloud.ir/ubuntu",
		Description: "ArvanCloud CDN mirror",
	},
	{
		Name:        "Runflare",
		URL:         "http://mirror-linux.runflare.com/ubuntu",
		Description: "Runflare mirror",
	},
}

var npmMirrors = []Mirror{
	{
		Name:        "Default",
		URL:         "https://registry.npmjs.org",
		Description: "npm official registry",
		IsDefault:   true,
	},
	{
		Name:        "Liara",
		URL:         "https://package-mirror.liara.ir/repository/npm/",
		Description: "Liara cloud npm mirror",
	},
	{
		Name:        "Runflare",
		URL:         "https://mirror-npm.runflare.com",
		Description: "Runflare npm mirror",
	},
	{
		Name:        "ParsPack",
		URL:         "https://mirror.abrha.net/repository/npm/",
		Description: "ParsPack (Abrha) npm mirror",
	},
}

var pipMirrors = []Mirror{
	{
		Name:        "Default",
		URL:         "https://pypi.org/simple",
		Description: "PyPI official",
		IsDefault:   true,
	},
	{
		Name:        "Liara",
		URL:         "https://package-mirror.liara.ir/repository/pypi/",
		Description: "Liara cloud PyPI mirror",
	},
	{
		Name:        "ParsPack",
		URL:         "https://mirror.abrha.net/repository/pypi/simple",
		Description: "ParsPack (Abrha) PyPI mirror",
	},
	{
		Name:        "Runflare",
		URL:         "https://mirror-pypi.runflare.com/simple",
		Description: "Runflare PyPI mirror",
	},
}

var dockerMirrors = []Mirror{
	{
		Name:        "Default",
		URL:         "https://registry-1.docker.io",
		Description: "Docker Hub official",
		IsDefault:   true,
	},
	{
		Name:        "Liara",
		URL:         "https://docker-mirror.liara.ir",
		Description: "Liara cloud Docker mirror",
	},
	{
		Name:        "ArvanCloud",
		URL:         "https://docker.arvancloud.ir",
		Description: "ArvanCloud Docker registry",
	},
	{
		Name:        "ParsPack",
		URL:         "https://docker.abrha.net",
		Description: "ParsPack (Abrha) Docker mirror",
	},
	{
		Name:        "Runflare",
		URL:         "https://mirror-docker.runflare.com",
		Description: "Runflare Docker mirror",
	},
}

var goMirrors = []Mirror{
	{
		Name:        "Default",
		URL:         "https://proxy.golang.org,direct",
		Description: "Go official proxy",
		IsDefault:   true,
	},
	{
		Name:        "ParsPack",
		URL:         "https://mirror.abrha.net/repository/go/,direct",
		Description: "ParsPack (Abrha) Go proxy",
	},
	{
		Name:        "Runflare",
		URL:         "https://mirror-go.runflare.com,direct",
		Description: "Runflare Go proxy",
	},
}

// ---- APT ----

const aptSourcesPath = "/etc/apt/sources.list"
const aptBackupPath = "/etc/apt/sources.list.bootup.bak"

func applyAPT(m Mirror) error {
	codename, err := getUbuntuCodename()
	if err != nil {
		return fmt.Errorf("cannot detect Ubuntu codename: %w", err)
	}

	utils.PrintInfo("Backing up " + aptSourcesPath + "...")
	utils.RunCommand("sudo", "cp", aptSourcesPath, aptBackupPath)

	securityURL := m.URL
	if su, ok := m.Meta["security_url"]; ok {
		securityURL = su
	}

	content := fmt.Sprintf(
		"deb %s/ %s main restricted universe multiverse\n"+
			"deb %s/ %s-updates main restricted universe multiverse\n"+
			"deb %s/ %s-backports main restricted universe multiverse\n"+
			"deb %s/ %s-security main restricted universe multiverse\n",
		m.URL, codename,
		m.URL, codename,
		m.URL, codename,
		securityURL, codename,
	)

	tmpFile, err := os.CreateTemp("", "sources.list.*")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		tmpFile.Close()
		return err
	}
	tmpFile.Close()

	if err := utils.RunCommand("sudo", "mv", tmpFile.Name(), aptSourcesPath); err != nil {
		return err
	}

	utils.PrintInfo("Running apt-get update...")
	return utils.RunCommand("sudo", "apt-get", "update")
}

func resetAPT() error {
	if _, err := os.Stat(aptBackupPath); os.IsNotExist(err) {
		return fmt.Errorf("no backup found at %s — run a mirror set first", aptBackupPath)
	}
	utils.PrintInfo("Restoring " + aptSourcesPath + " from backup...")
	if err := utils.RunCommand("sudo", "cp", aptBackupPath, aptSourcesPath); err != nil {
		return err
	}
	return utils.RunCommand("sudo", "apt-get", "update")
}

func statusAPT() string {
	data, err := os.ReadFile(aptSourcesPath)
	if err != nil {
		return "unknown"
	}
	content := string(data)
	for _, m := range aptMirrors {
		if m.IsDefault {
			continue
		}
		if strings.Contains(content, m.URL) {
			return m.Name
		}
	}
	return "default"
}

func getUbuntuCodename() (string, error) {
	out, err := exec.Command("lsb_release", "-sc").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// ---- npm ----

func applyNPM(m Mirror) error {
	return utils.RunCommand("npm", "config", "set", "registry", m.URL)
}

func resetNPM() error {
	return utils.RunCommand("npm", "config", "set", "registry", "https://registry.npmjs.org")
}

func statusNPM() string {
	out, err := exec.Command("npm", "config", "get", "registry").Output()
	if err != nil {
		return "npm not installed"
	}
	current := strings.TrimRight(strings.TrimSpace(string(out)), "/")
	for _, m := range npmMirrors {
		if m.IsDefault {
			continue
		}
		if current == strings.TrimRight(m.URL, "/") {
			return m.Name
		}
	}
	if current == "https://registry.npmjs.org" {
		return "default"
	}
	return "custom"
}

// ---- pip ----

const pipConfPath = "/etc/pip.conf"

func applyPIP(m Mirror) error {
	host := extractHost(m.URL)
	content := fmt.Sprintf("[global]\nindex-url = %s\ntrusted-host = %s\n", m.URL, host)

	if _, err := os.Stat(pipConfPath); err == nil {
		utils.RunCommand("sudo", "cp", pipConfPath, pipConfPath+".bootup.bak")
	}

	tmpFile, err := os.CreateTemp("", "pip.conf.*")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		tmpFile.Close()
		return err
	}
	tmpFile.Close()

	return utils.RunCommand("sudo", "mv", tmpFile.Name(), pipConfPath)
}

func resetPIP() error {
	backupPath := pipConfPath + ".bootup.bak"
	if _, err := os.Stat(backupPath); err == nil {
		return utils.RunCommand("sudo", "cp", backupPath, pipConfPath)
	}
	return utils.RunCommand("sudo", "rm", "-f", pipConfPath)
}

func statusPIP() string {
	paths := []string{pipConfPath}
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, home+"/.pip/pip.conf", home+"/.config/pip/pip.conf")
	}
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		content := string(data)
		for _, m := range pipMirrors {
			if m.IsDefault {
				continue
			}
			if strings.Contains(content, m.URL) {
				return m.Name
			}
		}
		if strings.Contains(content, "index-url") {
			return "custom"
		}
	}
	return "default"
}

// ---- Docker ----

const dockerDaemonPath = "/etc/docker/daemon.json"

func applyDocker(m Mirror) error {
	config := map[string]interface{}{}

	if data, err := os.ReadFile(dockerDaemonPath); err == nil {
		json.Unmarshal(data, &config)
		utils.RunCommand("sudo", "cp", dockerDaemonPath, dockerDaemonPath+".bootup.bak")
	}

	config["registry-mirrors"] = []string{m.URL}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	tmpFile, err := os.CreateTemp("", "daemon.json.*")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return err
	}
	tmpFile.Close()

	utils.RunCommand("sudo", "mkdir", "-p", "/etc/docker")
	if err := utils.RunCommand("sudo", "mv", tmpFile.Name(), dockerDaemonPath); err != nil {
		return err
	}

	utils.PrintInfo("Restarting Docker daemon...")
	return utils.RunCommand("sudo", "systemctl", "restart", "docker")
}

func resetDocker() error {
	backupPath := dockerDaemonPath + ".bootup.bak"
	if _, err := os.Stat(backupPath); err == nil {
		if err := utils.RunCommand("sudo", "cp", backupPath, dockerDaemonPath); err != nil {
			return err
		}
	} else {
		config := map[string]interface{}{}
		if data, err := os.ReadFile(dockerDaemonPath); err == nil {
			json.Unmarshal(data, &config)
		}
		delete(config, "registry-mirrors")

		if len(config) == 0 {
			utils.RunCommand("sudo", "rm", "-f", dockerDaemonPath)
		} else {
			data, _ := json.MarshalIndent(config, "", "  ")
			tmpFile, _ := os.CreateTemp("", "daemon.json.*")
			tmpFile.Write(data)
			tmpFile.Close()
			utils.RunCommand("sudo", "mv", tmpFile.Name(), dockerDaemonPath)
		}
	}

	utils.PrintInfo("Restarting Docker daemon...")
	return utils.RunCommand("sudo", "systemctl", "restart", "docker")
}

func statusDocker() string {
	data, err := os.ReadFile(dockerDaemonPath)
	if err != nil {
		return "default"
	}
	config := map[string]interface{}{}
	if err := json.Unmarshal(data, &config); err != nil {
		return "unknown"
	}
	mirrors, ok := config["registry-mirrors"]
	if !ok {
		return "default"
	}
	list, ok := mirrors.([]interface{})
	if !ok || len(list) == 0 {
		return "default"
	}
	url := fmt.Sprintf("%v", list[0])
	for _, m := range dockerMirrors {
		if m.IsDefault {
			continue
		}
		if m.URL == url {
			return m.Name
		}
	}
	return "custom"
}

// ---- Go ----

func applyGo(m Mirror) error {
	return utils.RunCommand("go", "env", "-w", "GOPROXY="+m.URL)
}

func resetGo() error {
	return utils.RunCommand("go", "env", "-w", "GOPROXY=https://proxy.golang.org,direct")
}

func statusGo() string {
	out, err := exec.Command("go", "env", "GOPROXY").Output()
	if err != nil {
		return "go not installed"
	}
	current := strings.TrimSpace(string(out))
	for _, m := range goMirrors {
		if m.IsDefault {
			continue
		}
		if current == m.URL {
			return m.Name
		}
	}
	if current == "https://proxy.golang.org,direct" || current == "" {
		return "default"
	}
	return "custom"
}

// ---- helpers ----

func extractHost(url string) string {
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	return strings.SplitN(url, "/", 2)[0]
}
