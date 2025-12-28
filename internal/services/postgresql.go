package services

import (
	"fmt"
	"github.com/amirkh8006/bootup-cli/internal/utils"
)

func InstallPostgreSQL() error {
	utils.PrintInfo("Installing PostgreSQL locally...")
	if err := utils.RunCommand("sudo", "apt-get", "install", "-y", "postgresql", "postgresql-contrib"); err != nil {
		return fmt.Errorf("failed to install PostgreSQL: %w", err)
	}

	utils.PrintInfo("Enabling PostgreSQL service...")
	if err := utils.RunCommand("sudo", "systemctl", "enable", "postgresql"); err != nil {
		return fmt.Errorf("failed to enable PostgreSQL service: %w", err)
	}

	utils.PrintInfo("Starting PostgreSQL service...")
	if err := utils.RunCommand("sudo", "systemctl", "start", "postgresql"); err != nil {
		return fmt.Errorf("failed to start PostgreSQL service: %w", err)
	}

	utils.PrintSuccess("PostgreSQL installed and started successfully!")
	return nil
}
