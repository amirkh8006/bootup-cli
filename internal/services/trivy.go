package services

import (
	"fmt"

	"github.com/amirkh8006/bootup-cli/internal/utils"
)

func InstallTrivy() error {
	utils.PrintInfo("Starting Trivy installation...")

	// Step 1: Install prerequisites
	utils.PrintInfo("Installing prerequisites...")
	if err := utils.RunCommand("sudo", "apt-get", "install", "-y", "wget", "gnupg"); err != nil {
		return fmt.Errorf("failed to install prerequisites: %w", err)
	}

	// Step 2: Add Trivy GPG key
	utils.PrintInfo("Adding Trivy GPG key...")
	keyCmd := "wget -qO - https://aquasecurity.github.io/trivy-repo/deb/public.key | gpg --dearmor | sudo tee /usr/share/keyrings/trivy.gpg > /dev/null"
	if err := utils.RunCommandShell(keyCmd); err != nil {
		return fmt.Errorf("failed to add Trivy GPG key: %w", err)
	}

	// Step 3: Add Trivy repository
	utils.PrintInfo("Adding Trivy repository...")
	repoCmd := `echo "deb [signed-by=/usr/share/keyrings/trivy.gpg] https://aquasecurity.github.io/trivy-repo/deb generic main" | sudo tee -a /etc/apt/sources.list.d/trivy.list`
	if err := utils.RunCommandShell(repoCmd); err != nil {
		return fmt.Errorf("failed to add Trivy repository: %w", err)
	}

	// Step 4: Update package index
	utils.PrintInfo("Updating package index...")
	if err := utils.RunCommand("sudo", "apt-get", "update"); err != nil {
		return fmt.Errorf("failed to update package index: %w", err)
	}

	// Step 5: Install Trivy
	utils.PrintInfo("Installing Trivy...")
	if err := utils.RunCommand("sudo", "apt-get", "install", "-y", "trivy"); err != nil {
		return fmt.Errorf("failed to install Trivy: %w", err)
	}

	// Step 6: Verify installation
	utils.PrintInfo("Verifying Trivy installation...")
	if err := utils.RunCommand("trivy", "version"); err != nil {
		return fmt.Errorf("Trivy installation verification failed: %w", err)
	}

	utils.PrintSuccess("Trivy installed successfully!")
	utils.PrintInfo("You can now use Trivy to scan for vulnerabilities:")
	utils.PrintInfo("  trivy image <image-name>     # Scan container images")
	utils.PrintInfo("  trivy fs <path>              # Scan filesystem")
	utils.PrintInfo("  trivy config <path>          # Scan configuration files")
	utils.PrintInfo("  trivy repo <repo-url>        # Scan remote repositories")

	return nil
}
