package services

import (
	"fmt"

	"github.com/amirkh8006/bootup-cli/internal/utils"
)

func InstallClickHouse() error {
	utils.PrintInfo("Installing ClickHouse locally...")

	// Install prerequisite packages
	utils.PrintInfo("Installing prerequisite packages...")
	if err := utils.RunCommand("sudo", "apt-get", "install", "-y", "apt-transport-https", "ca-certificates", "curl", "gnupg"); err != nil {
		return fmt.Errorf("failed to install prerequisite packages: %w", err)
	}

	// Download the ClickHouse GPG key and store it in the keyring
	utils.PrintInfo("Adding ClickHouse GPG key...")
	if err := utils.RunCommand("bash", "-c", "curl -fsSL 'https://packages.clickhouse.com/rpm/lts/repodata/repomd.xml.key' | sudo gpg --dearmor -o /usr/share/keyrings/clickhouse-keyring.gpg"); err != nil {
		return fmt.Errorf("failed to add ClickHouse GPG key: %w", err)
	}

	// Get the system architecture and add the ClickHouse repository
	utils.PrintInfo("Adding ClickHouse repository...")
	if err := utils.RunCommand("bash", "-c", "ARCH=$(dpkg --print-architecture) && echo \"deb [signed-by=/usr/share/keyrings/clickhouse-keyring.gpg arch=${ARCH}] https://packages.clickhouse.com/deb stable main\" | sudo tee /etc/apt/sources.list.d/clickhouse.list"); err != nil {
		return fmt.Errorf("failed to add ClickHouse repository: %w", err)
	}

	// Update package list
	utils.PrintInfo("Updating package list...")
	if err := utils.RunCommand("sudo", "apt-get", "update"); err != nil {
		return fmt.Errorf("failed to update packages: %w", err)
	}

	// Install ClickHouse server and client
	utils.PrintInfo("Installing ClickHouse server and client...")
	if err := utils.RunCommand("sudo", "apt-get", "install", "-y", "clickhouse-server", "clickhouse-client"); err != nil {
		return fmt.Errorf("failed to install ClickHouse: %w", err)
	}

	// Enable ClickHouse service
	utils.PrintInfo("Enabling ClickHouse service...")
	if err := utils.RunCommand("sudo", "systemctl", "enable", "clickhouse-server"); err != nil {
		return fmt.Errorf("failed to enable ClickHouse service: %w", err)
	}

	// Start ClickHouse service
	utils.PrintInfo("Starting ClickHouse service...")
	if err := utils.RunCommand("sudo", "systemctl", "start", "clickhouse-server"); err != nil {
		return fmt.Errorf("failed to start ClickHouse service: %w", err)
	}

	utils.PrintSuccess("ClickHouse installed and started successfully!")
	utils.PrintInfo("You can connect to ClickHouse using: clickhouse-client")
	utils.PrintInfo("If you set up a password, use: clickhouse-client --password")
	return nil
}
