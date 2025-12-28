package services

import (
	"fmt"

	"github.com/amirkh8006/bootup-cli/internal/utils"
)

func InstallRedis() error {
	utils.PrintInfo("Installing Redis locally...")
	if err := utils.RunCommand("sudo", "apt-get", "install", "-y", "redis-server"); err != nil {
		return fmt.Errorf("failed to install Redis: %w", err)
	}

	utils.PrintInfo("Enabling Redis service...")
	if err := utils.RunCommand("sudo", "systemctl", "enable", "redis-server"); err != nil {
		return fmt.Errorf("failed to enable Redis service: %w", err)
	}

	utils.PrintInfo("Starting Redis service...")
	if err := utils.RunCommand("sudo", "systemctl", "start", "redis-server"); err != nil {
		return fmt.Errorf("failed to start Redis service: %w", err)
	}

	utils.PrintSuccess("Redis installed and started successfully!")
	return nil
}
