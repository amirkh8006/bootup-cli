package services

import (
	"fmt"
	"github.com/amirkh8006/bootup-cli/internal/utils"
)

func InstallPrometheus() error {
	utils.PrintInfo("Installing Prometheus...")

	// Define version and directories
	prometheusVersion := "3.0.1"
	prometheusUser := "prometheus"
	prometheusDir := "/opt/prometheus"
	prometheusDataDir := "/var/lib/prometheus"
	prometheusConfigFile := "/etc/prometheus/prometheus.yml"
	prometheusTarball := "/tmp/prometheus.tar.gz"

	// Create prometheus user if not exists
	utils.PrintInfo("Creating prometheus user...")
	checkUserCmd := fmt.Sprintf("id %s", prometheusUser)
	if err := utils.RunCommandShell(checkUserCmd); err != nil {
		// User doesn't exist, create it
		createUserCmd := fmt.Sprintf("sudo useradd --no-create-home --shell /usr/sbin/nologin %s", prometheusUser)
		if err := utils.RunCommandShell(createUserCmd); err != nil {
			return fmt.Errorf("failed to create prometheus user: %w", err)
		}
	}

	// Download Prometheus
	utils.PrintInfo(fmt.Sprintf("Downloading Prometheus %s...", prometheusVersion))
	downloadUrl := fmt.Sprintf("https://github.com/prometheus/prometheus/releases/download/v%s/prometheus-%s.linux-amd64.tar.gz", prometheusVersion, prometheusVersion)
	downloadCmd := fmt.Sprintf("wget '%s' -O %s", downloadUrl, prometheusTarball)
	if err := utils.RunCommandShell(downloadCmd); err != nil {
		return fmt.Errorf("failed to download Prometheus: %w", err)
	}

	// Create and extract to directory
	utils.PrintInfo("Extracting Prometheus...")
	if err := utils.RunCommandShell(fmt.Sprintf("sudo mkdir -p %s", prometheusDir)); err != nil {
		return fmt.Errorf("failed to create prometheus directory: %w", err)
	}
	
	extractCmd := fmt.Sprintf("sudo tar -xzf %s --strip-components=1 -C %s", prometheusTarball, prometheusDir)
	if err := utils.RunCommandShell(extractCmd); err != nil {
		return fmt.Errorf("failed to extract Prometheus: %w", err)
	}

	// Create data and config directories
	utils.PrintInfo("Creating configuration directories...")
	if err := utils.RunCommandShell(fmt.Sprintf("sudo mkdir -p %s /etc/prometheus", prometheusDataDir)); err != nil {
		return fmt.Errorf("failed to create prometheus data directory: %w", err)
	}

	// Create default config
	utils.PrintInfo("Creating default configuration...")
	configContent := `global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']`
	
	createConfigCmd := fmt.Sprintf("sudo tee %s > /dev/null <<'EOL'\n%s\nEOL", prometheusConfigFile, configContent)
	if err := utils.RunCommandShell(createConfigCmd); err != nil {
		return fmt.Errorf("failed to create prometheus config: %w", err)
	}

	// Fix permissions
	utils.PrintInfo("Setting permissions...")
	chownCmd := fmt.Sprintf("sudo chown -R %s:%s %s %s /etc/prometheus", prometheusUser, prometheusUser, prometheusDir, prometheusDataDir)
	if err := utils.RunCommandShell(chownCmd); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Create systemd service
	utils.PrintInfo("Creating systemd service...")
	serviceContent := fmt.Sprintf(`[Unit]
Description=Prometheus Monitoring
Wants=network-online.target
After=network-online.target

[Service]
User=%s
Group=%s
Type=simple
ExecStart=%s/prometheus \
  --config.file=%s \
  --storage.tsdb.path=%s \
  --web.listen-address=:9090 \
  --web.external-url=https://prometheus.khanetalaa.ir/

Restart=always

[Install]
WantedBy=multi-user.target`, prometheusUser, prometheusUser, prometheusDir, prometheusConfigFile, prometheusDataDir)
	
	createServiceCmd := fmt.Sprintf("sudo tee /etc/systemd/system/prometheus.service > /dev/null <<'EOL'\n%s\nEOL", serviceContent)
	if err := utils.RunCommandShell(createServiceCmd); err != nil {
		return fmt.Errorf("failed to create systemd service: %w", err)
	}

	// Enable and start service
	utils.PrintInfo("Starting Prometheus service...")
	if err := utils.RunCommand("sudo", "systemctl", "daemon-reload"); err != nil {
		return fmt.Errorf("failed to reload systemd: %w", err)
	}
	
	if err := utils.RunCommand("sudo", "systemctl", "enable", "prometheus"); err != nil {
		return fmt.Errorf("failed to enable prometheus service: %w", err)
	}
	
	if err := utils.RunCommand("sudo", "systemctl", "start", "prometheus"); err != nil {
		return fmt.Errorf("failed to start prometheus service: %w", err)
	}

	// Clean up downloaded tarball
	utils.RunCommandShell(fmt.Sprintf("rm -f %s", prometheusTarball))

	utils.PrintSuccess("Prometheus installed and running!")
	utils.PrintInfo("Prometheus is accessible at http://localhost:9090")
	
	return nil
}
