package services

import (
	"fmt"

	"github.com/amirkh8006/bootup-cli/internal/utils"
)

func InstallCaddy() error {
	utils.PrintInfo("Installing required packages...")
	if err := utils.RunCommand("sudo", "apt", "install", "-y", "debian-keyring", "debian-archive-keyring", "apt-transport-https", "curl"); err != nil {
		return fmt.Errorf("failed to install required packages: %w", err)
	}

	utils.PrintInfo("Adding Caddy GPG key...")
	if err := utils.RunCommand("bash", "-c", "curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg"); err != nil {
		return fmt.Errorf("failed to add Caddy GPG key: %w", err)
	}

	utils.PrintInfo("Adding Caddy repository...")
	if err := utils.RunCommand("bash", "-c", "curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list"); err != nil {
		return fmt.Errorf("failed to add Caddy repository: %w", err)
	}

	utils.PrintInfo("Setting repository permissions...")
	if err := utils.RunCommand("sudo", "chmod", "o+r", "/usr/share/keyrings/caddy-stable-archive-keyring.gpg"); err != nil {
		return fmt.Errorf("failed to set GPG key permissions: %w", err)
	}

	if err := utils.RunCommand("sudo", "chmod", "o+r", "/etc/apt/sources.list.d/caddy-stable.list"); err != nil {
		return fmt.Errorf("failed to set repository permissions: %w", err)
	}

	utils.PrintInfo("Updating package list...")
	if err := utils.RunCommand("sudo", "apt", "update"); err != nil {
		return fmt.Errorf("failed to update packages: %w", err)
	}

	utils.PrintInfo("Installing Caddy...")
	if err := utils.RunCommand("sudo", "apt", "install", "-y", "caddy"); err != nil {
		return fmt.Errorf("failed to install Caddy: %w", err)
	}

	utils.PrintSuccess("Caddy installed successfully!")
	utils.PrintInfo("Caddy is now available. You can start it with: sudo systemctl start caddy")
	utils.PrintInfo("To enable it on boot: sudo systemctl enable caddy")
	utils.PrintInfo("Default configuration file: /etc/caddy/Caddyfile")

	return nil
}
