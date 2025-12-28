package services

import (
	"fmt"

	"github.com/amirkh8006/bootup-cli/internal/utils"
)

func InstallGrafana() error {
	utils.PrintInfo("Installing Grafana...")

	// Install prerequisites
	utils.PrintInfo("Installing prerequisites...")
	if err := utils.RunCommand("sudo", "apt-get", "install", "-y", "apt-transport-https", "software-properties-common", "wget"); err != nil {
		return fmt.Errorf("failed to install prerequisites for Grafana: %w", err)
	}

	// Create keyrings directory
	utils.PrintInfo("Creating keyrings directory...")
	if err := utils.RunCommand("sudo", "mkdir", "-p", "/etc/apt/keyrings/"); err != nil {
		return fmt.Errorf("failed to create keyrings directory: %w", err)
	}

	// Add Grafana GPG key
	utils.PrintInfo("Adding Grafana GPG key...")
	gpgKeyCmd := "wget -q -O - https://apt.grafana.com/gpg.key | gpg --dearmor | sudo tee /etc/apt/keyrings/grafana.gpg > /dev/null"
	if err := utils.RunCommandShell(gpgKeyCmd); err != nil {
		return fmt.Errorf("failed to add Grafana GPG key: %w", err)
	}

	// Add Grafana repository
	utils.PrintInfo("Adding Grafana repository...")
	repoCmd := "echo 'deb [signed-by=/etc/apt/keyrings/grafana.gpg] https://apt.grafana.com stable main' | sudo tee -a /etc/apt/sources.list.d/grafana.list"
	if err := utils.RunCommandShell(repoCmd); err != nil {
		return fmt.Errorf("failed to add Grafana repository: %w", err)
	}

	// Update package lists
	utils.PrintInfo("Updating package lists...")
	if err := utils.RunCommand("sudo", "apt-get", "update", "-y"); err != nil {
		return fmt.Errorf("failed to update package lists after adding Grafana repo: %w", err)
	}

	// Install Grafana
	utils.PrintInfo("Installing Grafana package...")
	if err := utils.RunCommand("sudo", "apt-get", "install", "-y", "grafana"); err != nil {
		return fmt.Errorf("failed to install Grafana: %w", err)
	}

	// Reload systemd daemon
	utils.PrintInfo("Reloading systemd daemon...")
	if err := utils.RunCommand("sudo", "systemctl", "daemon-reload"); err != nil {
		return fmt.Errorf("failed to reload systemd daemon: %w", err)
	}

	// Enable Grafana service
	utils.PrintInfo("Enabling Grafana service...")
	if err := utils.RunCommand("sudo", "systemctl", "enable", "grafana-server"); err != nil {
		return fmt.Errorf("failed to enable Grafana service: %w", err)
	}

	// Start Grafana service
	utils.PrintInfo("Starting Grafana service...")
	if err := utils.RunCommand("sudo", "systemctl", "start", "grafana-server"); err != nil {
		return fmt.Errorf("failed to start Grafana service: %w", err)
	}

	utils.PrintSuccess("Grafana installed and running!")
	utils.PrintInfo("Grafana is accessible at http://localhost:3000")
	utils.PrintInfo("Default login: admin/admin")

	return nil
}
