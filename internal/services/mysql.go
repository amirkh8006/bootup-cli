package services

import (
	"fmt"

	"github.com/amirkh8006/bootup-cli/internal/utils"
)

func InstallMySQL() error {
	utils.PrintInfo("Installing MySQL Server locally...")

	// Install MySQL server
	if err := utils.RunCommand("sudo", "apt-get", "update"); err != nil {
		return fmt.Errorf("failed to update package list: %w", err)
	}

	if err := utils.RunCommand("sudo", "apt-get", "install", "-y", "mysql-server"); err != nil {
		return fmt.Errorf("failed to install MySQL server: %w", err)
	}

	utils.PrintInfo("Enabling MySQL service...")
	if err := utils.RunCommand("sudo", "systemctl", "enable", "mysql"); err != nil {
		return fmt.Errorf("failed to enable MySQL service: %w", err)
	}

	utils.PrintInfo("Starting MySQL service...")
	if err := utils.RunCommand("sudo", "systemctl", "start", "mysql"); err != nil {
		return fmt.Errorf("failed to start MySQL service: %w", err)
	}

	utils.PrintInfo("Securing MySQL installation...")
	utils.PrintInfo("Note: Run 'sudo mysql_secure_installation' manually to complete the secure setup")

	utils.PrintSuccess("MySQL installed and started successfully!")
	utils.PrintInfo("Default connection: mysql -u root -p")
	return nil
}
