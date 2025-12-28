package services

import (
	"fmt"

	"github.com/amirkh8006/bootup-cli/internal/utils"
)

func InstallMongoDB() error {
	utils.PrintInfo("Installing MongoDB locally...")

	// Add MongoDB GPG key
	utils.PrintInfo("Adding MongoDB GPG key...")
	if err := utils.RunCommand("bash", "-c", "curl -fsSL https://www.mongodb.org/static/pgp/server-8.0.asc | sudo gpg -o /usr/share/keyrings/mongodb-server-8.0.gpg --dearmor"); err != nil {
		return fmt.Errorf("failed to add MongoDB GPG key: %w", err)
	}

	// Add MongoDB repository
	utils.PrintInfo("Adding MongoDB repository...")
	if err := utils.RunCommand("bash", "-c", "echo \"deb [ arch=amd64,arm64 signed-by=/usr/share/keyrings/mongodb-server-8.0.gpg ] https://repo.mongodb.org/apt/ubuntu noble/mongodb-org/8.2 multiverse\" | sudo tee /etc/apt/sources.list.d/mongodb-org-8.2.list"); err != nil {
		return fmt.Errorf("failed to add MongoDB repository: %w", err)
	}

	// Update package list
	utils.PrintInfo("Updating package list...")
	if err := utils.RunCommand("sudo", "apt-get", "update", "-y"); err != nil {
		return fmt.Errorf("failed to update packages: %w", err)
	}

	// Install MongoDB
	utils.PrintInfo("Installing MongoDB...")
	if err := utils.RunCommand("sudo", "apt-get", "install", "-y", "mongodb-org"); err != nil {
		return fmt.Errorf("failed to install MongoDB: %w", err)
	}

	// Enable MongoDB service
	utils.PrintInfo("Enabling MongoDB service...")
	if err := utils.RunCommand("sudo", "systemctl", "enable", "mongod"); err != nil {
		return fmt.Errorf("failed to enable MongoDB service: %w", err)
	}

	// Start MongoDB service
	utils.PrintInfo("Starting MongoDB service...")
	if err := utils.RunCommand("sudo", "systemctl", "start", "mongod"); err != nil {
		return fmt.Errorf("failed to start MongoDB service: %w", err)
	}

	utils.PrintSuccess("MongoDB installed and started successfully!")
	return nil
}
