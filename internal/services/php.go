package services

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/amirkh8006/bootup-cli/internal/utils"
)

func InstallPHP() error {
	utils.PrintInfo("Installing PHP...")

	// Check if running on supported OS
	if runtime.GOOS != "linux" {
		return fmt.Errorf("PHP installation is currently only supported on Linux")
	}

	// For now, assume Ubuntu/Debian (like Docker service)
	// Can be extended later for other distributions
	return installPHPDebian()
}

func installPHPDebian() error {
	utils.PrintInfo("Installing PHP on Ubuntu/Debian...")

	// Update package lists
	if err := utils.RunCommand("sudo", "apt", "update"); err != nil {
		return fmt.Errorf("failed to update package lists: %w", err)
	}

	// Add Ondřej Surý's PPA for additional PHP versions (Ubuntu only)
	utils.PrintInfo("Adding PHP PPA for more version options...")
	if err := utils.RunCommand("sudo", "apt", "install", "-y", "software-properties-common"); err != nil {
		utils.PrintWarning("Failed to install software-properties-common, continuing with default repositories")
	} else {
		if err := utils.RunCommand("sudo", "add-apt-repository", "-y", "ppa:ondrej/php"); err != nil {
			utils.PrintWarning("Failed to add PHP PPA, using default repositories")
		} else {
			// Update package lists after adding PPA
			if err := utils.RunCommand("sudo", "apt", "update"); err != nil {
				return fmt.Errorf("failed to update package lists after adding PPA: %w", err)
			}
		}
	}

	// Display available PHP versions
	versions := []string{
		"php8.3",
		"php8.2",
		"php8.1",
		"php8.0",
		"php7.4",
	}

	selectedVersion, err := displayAndSelectPHPVersion(versions)
	if err != nil {
		return fmt.Errorf("failed to select version: %w", err)
	}

	// Install PHP and common extensions
	packages := []string{
		selectedVersion,
		selectedVersion + "-cli",
		selectedVersion + "-common",
		selectedVersion + "-curl",
		selectedVersion + "-gd",
		selectedVersion + "-mbstring",
		selectedVersion + "-mysql",
		selectedVersion + "-xml",
		selectedVersion + "-zip",
		selectedVersion + "-fpm",
		selectedVersion + "-opcache",
		selectedVersion + "-intl",
		selectedVersion + "-bcmath",
	}

	utils.PrintInfo(fmt.Sprintf("Installing %s and common extensions...", selectedVersion))
	args := append([]string{"apt", "install", "-y"}, packages...)
	if err := utils.RunCommand("sudo", args...); err != nil {
		return fmt.Errorf("failed to install PHP: %w", err)
	}

	// Install Composer
	return installComposer()
}

func displayAndSelectPHPVersion(versions []string) (string, error) {
	if len(versions) == 0 {
		return "", fmt.Errorf("no PHP versions available")
	}

	utils.PrintInfo("Available PHP versions:")
	for i, version := range versions {
		fmt.Printf("%d. %s\n", i+1, version)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Select a version (default: 1): ")
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return versions[0], nil
	}

	choice, err := strconv.Atoi(input)
	if err != nil || choice < 1 || choice > len(versions) {
		return "", fmt.Errorf("invalid selection")
	}

	return versions[choice-1], nil
}

func installComposer() error {
	utils.PrintInfo("Installing Composer...")

	// Download Composer installer
	if err := utils.RunCommand("curl", "-sS", "https://getcomposer.org/installer", "-o", "composer-setup.php"); err != nil {
		return fmt.Errorf("failed to download Composer installer: %w", err)
	}

	// Install Composer
	if err := utils.RunCommand("sudo", "php", "composer-setup.php", "--install-dir=/usr/local/bin", "--filename=composer"); err != nil {
		return fmt.Errorf("failed to install Composer: %w", err)
	}

	// Clean up installer
	if err := utils.RunCommand("rm", "composer-setup.php"); err != nil {
		utils.PrintWarning("Failed to remove Composer installer")
	}

	utils.PrintSuccess("PHP and Composer installed successfully!")
	utils.PrintInfo("You can verify the installation with:")
	utils.PrintInfo("  php --version")
	utils.PrintInfo("  composer --version")

	return nil
}
