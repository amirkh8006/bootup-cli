package services

import (
	"fmt"

	"github.com/amirkh8006/bootup-cli/internal/utils"
)

func InstallDocker() error {
	utils.PrintInfo("Starting Docker installation...")

	// Step 1: Remove old conflicting packages
	utils.PrintInfo("Removing any conflicting packages...")
	if err := utils.RunCommand("sudo", "apt", "remove", "docker.io", "docker-compose", "docker-compose-v2", "docker-doc", "podman-docker", "containerd", "runc"); err != nil {
		utils.PrintWarning("No conflicting packages found to remove (this is normal)")
	}

	// Step 2: Update package index
	utils.PrintInfo("Updating package index...")
	if err := utils.RunCommand("sudo", "apt", "update"); err != nil {
		return fmt.Errorf("failed to update package index: %w", err)
	}

	// Step 3: Install prerequisites
	utils.PrintInfo("Installing prerequisites...")
	if err := utils.RunCommand("sudo", "apt", "install", "-y", "ca-certificates", "curl"); err != nil {
		return fmt.Errorf("failed to install prerequisites: %w", err)
	}

	// Step 4: Create keyrings directory
	utils.PrintInfo("Setting up Docker's GPG key...")
	if err := utils.RunCommand("sudo", "install", "-m", "0755", "-d", "/etc/apt/keyrings"); err != nil {
		return fmt.Errorf("failed to create keyrings directory: %w", err)
	}

	// Step 5: Add Docker's GPG key
	if err := utils.RunCommand("sudo", "curl", "-fsSL", "https://download.docker.com/linux/ubuntu/gpg", "-o", "/etc/apt/keyrings/docker.asc"); err != nil {
		return fmt.Errorf("failed to download Docker GPG key: %w", err)
	}

	if err := utils.RunCommand("sudo", "chmod", "a+r", "/etc/apt/keyrings/docker.asc"); err != nil {
		return fmt.Errorf("failed to set GPG key permissions: %w", err)
	}

	// Step 6: Add Docker repository
	utils.PrintInfo("Adding Docker repository...")
	dockerRepoCmd := `echo "Types: deb
URIs: https://download.docker.com/linux/ubuntu
Suites: $(. /etc/os-release && echo "${UBUNTU_CODENAME:-$VERSION_CODENAME}")
Components: stable
Signed-By: /etc/apt/keyrings/docker.asc" | sudo tee /etc/apt/sources.list.d/docker.sources > /dev/null`

	if err := utils.RunCommandShell(dockerRepoCmd); err != nil {
		return fmt.Errorf("failed to add Docker repository: %w", err)
	}

	// Step 7: Update package index again
	utils.PrintInfo("Updating package index with Docker repository...")
	if err := utils.RunCommand("sudo", "apt", "update"); err != nil {
		return fmt.Errorf("failed to update package index: %w", err)
	}

	// Step 8: Install Docker packages
	utils.PrintInfo("Installing Docker packages...")
	if err := utils.RunCommand("sudo", "apt", "install", "-y", "docker-ce", "docker-ce-cli", "containerd.io", "docker-buildx-plugin", "docker-compose-plugin"); err != nil {
		return fmt.Errorf("failed to install Docker packages: %w", err)
	}

	// Step 9: Start and enable Docker service
	utils.PrintInfo("Starting Docker service...")
	if err := utils.RunCommand("sudo", "systemctl", "start", "docker"); err != nil {
		return fmt.Errorf("failed to start Docker service: %w", err)
	}

	utils.PrintInfo("Enabling Docker service...")
	if err := utils.RunCommand("sudo", "systemctl", "enable", "docker"); err != nil {
		return fmt.Errorf("failed to enable Docker service: %w", err)
	}

	// Step 10: Verify installation
	utils.PrintInfo("Verifying Docker installation...")
	if err := utils.RunCommand("sudo", "docker", "run", "hello-world"); err != nil {
		return fmt.Errorf("Docker installation verification failed: %w", err)
	}

	// Step 11: Add current user to docker group (optional)
	utils.PrintInfo("Adding current user to docker group...")
	if err := utils.RunCommandShell("sudo usermod -aG docker $USER"); err != nil {
		utils.PrintWarning("Failed to add user to docker group - you may need to run Docker commands with sudo")
	} else {
		utils.PrintInfo("User added to docker group. You may need to log out and back in for changes to take effect.")
	}

	utils.PrintSuccess("Docker installed successfully!")
	utils.PrintInfo("To use Docker without sudo, please log out and back in, or run: newgrp docker")
	return nil
}
