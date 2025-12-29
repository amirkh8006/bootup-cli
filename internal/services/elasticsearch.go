package services

import (
	"fmt"

	"github.com/amirkh8006/bootup-cli/internal/utils"
)

func InstallElasticsearch() error {
	utils.PrintInfo("Installing Elasticsearch locally...")

	// Step 1: Import the Elasticsearch PGP key
	utils.PrintInfo("Adding Elasticsearch GPG key...")
	if err := utils.RunCommand("bash", "-c", "wget -qO - https://artifacts.elastic.co/GPG-KEY-elasticsearch | sudo gpg --dearmor -o /usr/share/keyrings/elasticsearch-keyring.gpg"); err != nil {
		return fmt.Errorf("failed to add Elasticsearch GPG key: %w", err)
	}

	// Step 2: Install apt-transport-https if needed (for Debian)
	utils.PrintInfo("Installing apt-transport-https...")
	if err := utils.RunCommand("sudo", "apt-get", "install", "-y", "apt-transport-https"); err != nil {
		return fmt.Errorf("failed to install apt-transport-https: %w", err)
	}

	// Step 3: Add Elasticsearch repository
	utils.PrintInfo("Adding Elasticsearch repository...")
	if err := utils.RunCommand("bash", "-c", "echo \"deb [signed-by=/usr/share/keyrings/elasticsearch-keyring.gpg] https://artifacts.elastic.co/packages/9.x/apt stable main\" | sudo tee /etc/apt/sources.list.d/elastic-9.x.list"); err != nil {
		return fmt.Errorf("failed to add Elasticsearch repository: %w", err)
	}

	// Step 4: Update package list
	utils.PrintInfo("Updating package list...")
	if err := utils.RunCommand("sudo", "apt-get", "update", "-y"); err != nil {
		return fmt.Errorf("failed to update packages: %w", err)
	}

	// Step 5: Install Elasticsearch
	utils.PrintInfo("Installing Elasticsearch...")
	if err := utils.RunCommand("sudo", "apt-get", "install", "-y", "elasticsearch"); err != nil {
		return fmt.Errorf("failed to install Elasticsearch: %w", err)
	}

	// Step 6: Enable Elasticsearch service
	utils.PrintInfo("Enabling Elasticsearch service...")
	if err := utils.RunCommand("sudo", "/bin/systemctl", "daemon-reload"); err != nil {
		return fmt.Errorf("failed to reload systemd daemon: %w", err)
	}

	if err := utils.RunCommand("sudo", "/bin/systemctl", "enable", "elasticsearch.service"); err != nil {
		return fmt.Errorf("failed to enable Elasticsearch service: %w", err)
	}

	// Step 7: Start Elasticsearch service
	utils.PrintInfo("Starting Elasticsearch service...")
	if err := utils.RunCommand("sudo", "systemctl", "start", "elasticsearch.service"); err != nil {
		return fmt.Errorf("failed to start Elasticsearch service: %w", err)
	}

	utils.PrintSuccess("Elasticsearch installed and started successfully!")
	utils.PrintInfo("Note: Elasticsearch runs on localhost:9200 by default.")
	utils.PrintInfo("Security is auto-configured. Use 'sudo /usr/share/elasticsearch/bin/elasticsearch-reset-password -u elastic' to reset the elastic user password.")
	utils.PrintInfo("Check status with: sudo systemctl status elasticsearch.service")

	return nil
}
