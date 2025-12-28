package services

import (
	"fmt"

	"github.com/amirkh8006/bootup-cli/internal/utils"
)

func InstallAlertmanager() error {
	utils.PrintInfo("Installing Prometheus Alertmanager...")

	// Define version and directories
	alertmanagerVersion := "0.28.1"
	alertmanagerUser := "prometheus"
	alertmanagerDir := "/opt/alertmanager"
	alertmanagerDataDir := "/var/lib/alertmanager"
	alertmanagerConfigFile := "/etc/alertmanager/alertmanager.yml"
	alertmanagerTarball := "/tmp/alertmanager.tar.gz"

	// Create prometheus user if not exists (reuse prometheus user)
	utils.PrintInfo("Checking prometheus user...")
	checkUserCmd := fmt.Sprintf("id %s", alertmanagerUser)
	if err := utils.RunCommandShell(checkUserCmd); err != nil {
		utils.PrintInfo("Creating prometheus user...")
		createUserCmd := fmt.Sprintf("sudo useradd --no-create-home --shell /usr/sbin/nologin %s", alertmanagerUser)
		if err := utils.RunCommandShell(createUserCmd); err != nil {
			return fmt.Errorf("failed to create prometheus user: %w", err)
		}
	}

	// Download Alertmanager
	utils.PrintInfo(fmt.Sprintf("Downloading Alertmanager %s...", alertmanagerVersion))
	downloadUrl := fmt.Sprintf("https://github.com/prometheus/alertmanager/releases/download/v%s/alertmanager-%s.linux-amd64.tar.gz", alertmanagerVersion, alertmanagerVersion)
	downloadCmd := fmt.Sprintf("wget '%s' -O %s", downloadUrl, alertmanagerTarball)
	if err := utils.RunCommandShell(downloadCmd); err != nil {
		return fmt.Errorf("failed to download Alertmanager: %w", err)
	}

	// Create and extract to directory
	utils.PrintInfo("Extracting Alertmanager...")
	if err := utils.RunCommandShell(fmt.Sprintf("sudo mkdir -p %s", alertmanagerDir)); err != nil {
		return fmt.Errorf("failed to create alertmanager directory: %w", err)
	}

	extractCmd := fmt.Sprintf("sudo tar -xzf %s --strip-components=1 -C %s", alertmanagerTarball, alertmanagerDir)
	if err := utils.RunCommandShell(extractCmd); err != nil {
		return fmt.Errorf("failed to extract Alertmanager: %w", err)
	}

	// Create data and config directories
	utils.PrintInfo("Creating configuration directories...")
	if err := utils.RunCommandShell(fmt.Sprintf("sudo mkdir -p %s /etc/alertmanager", alertmanagerDataDir)); err != nil {
		return fmt.Errorf("failed to create alertmanager directories: %w", err)
	}

	// Create default configuration
	utils.PrintInfo("Creating default configuration...")
	configContent := `global:
  resolve_timeout: 5m

route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'web.hook'

receivers:
  - name: 'web.hook'
    webhook_configs:
      - url: 'http://127.0.0.1:5001/'

inhibit_rules:
  - source_match:
      severity: 'critical'
    target_match:
      severity: 'warning'
    equal: ['alertname', 'dev', 'instance']`

	createConfigCmd := fmt.Sprintf("sudo tee %s > /dev/null <<'EOL'\n%s\nEOL", alertmanagerConfigFile, configContent)
	if err := utils.RunCommandShell(createConfigCmd); err != nil {
		return fmt.Errorf("failed to create alertmanager config: %w", err)
	}

	// Create systemd service
	utils.PrintInfo("Creating systemd service...")
	serviceContent := fmt.Sprintf(`[Unit]
Description=Prometheus Alertmanager
Wants=network-online.target
After=network-online.target

[Service]
User=%s
Group=%s
Type=simple
ExecStart=%s/alertmanager \
  --config.file=%s \
  --storage.path=%s \
  --cluster.listen-address="" \
  --web.listen-address=:9094

Restart=always

[Install]
WantedBy=multi-user.target`, alertmanagerUser, alertmanagerUser, alertmanagerDir, alertmanagerConfigFile, alertmanagerDataDir)

	createServiceCmd := fmt.Sprintf("sudo tee /etc/systemd/system/alertmanager.service > /dev/null <<'EOL'\n%s\nEOL", serviceContent)
	if err := utils.RunCommandShell(createServiceCmd); err != nil {
		return fmt.Errorf("failed to create systemd service: %w", err)
	}

	// Fix permissions
	utils.PrintInfo("Setting permissions...")
	chownCmd := fmt.Sprintf("sudo chown -R %s:%s %s %s /etc/alertmanager", alertmanagerUser, alertmanagerUser, alertmanagerDir, alertmanagerDataDir)
	if err := utils.RunCommandShell(chownCmd); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Enable and start service
	utils.PrintInfo("Starting Alertmanager service...")
	if err := utils.RunCommand("sudo", "systemctl", "daemon-reload"); err != nil {
		return fmt.Errorf("failed to reload systemd: %w", err)
	}

	if err := utils.RunCommand("sudo", "systemctl", "enable", "alertmanager"); err != nil {
		return fmt.Errorf("failed to enable alertmanager service: %w", err)
	}

	if err := utils.RunCommand("sudo", "systemctl", "start", "alertmanager"); err != nil {
		return fmt.Errorf("failed to start alertmanager service: %w", err)
	}

	// Clean up downloaded tarball
	utils.RunCommandShell(fmt.Sprintf("rm -f %s", alertmanagerTarball))

	utils.PrintSuccess("Prometheus Alertmanager installed and running!")
	utils.PrintInfo("Alertmanager is accessible at http://localhost:9094")

	return nil
}
