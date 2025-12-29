package services

import (
	"fmt"
	"os"

	"github.com/amirkh8006/bootup-cli/internal/utils"
)

func InstallRustFS() error {
	// Check if RustFS is already installed
	if isRustFSInstalled() {
		utils.PrintInfo("RustFS is already installed")
		utils.PrintInfo("Attempting to upgrade RustFS...")
		return upgradeRustFS()
	}

	utils.PrintInfo("Downloading RustFS installer...")

	// Download the installation script
	if err := utils.RunCommand("curl", "-O", "https://rustfs.com/install_rustfs.sh"); err != nil {
		return fmt.Errorf("failed to download RustFS installer: %w", err)
	}

	utils.PrintInfo("Installing RustFS...")

	// Make the script executable
	if err := utils.RunCommand("chmod", "+x", "install_rustfs.sh"); err != nil {
		return fmt.Errorf("failed to make installer executable: %w", err)
	}

	// Run the script with sudo and automatically provide responses:
	// - Choice: 1 (Install)
	// - Port: 9000 (default)
	// - Console Port: 9001 (default)
	// - Data Directory: /data/rustfs0 (default)
	installInput := "1\n9000\n9001\n/data/rustfs0\n"
	if err := utils.RunCommandShell(fmt.Sprintf("echo -e '%s' | sudo bash install_rustfs.sh", installInput)); err != nil {
		return fmt.Errorf("failed to install RustFS: %w", err)
	}

	// Clean up the installer script
	if err := os.Remove("install_rustfs.sh"); err != nil {
		utils.PrintWarning(fmt.Sprintf("Failed to clean up installer script: %v", err))
	}

	utils.PrintSuccess("RustFS installed successfully!")
	return nil
}

func isRustFSInstalled() bool {
	// Check if the RustFS binary exists at the expected location
	if _, err := os.Stat("/usr/local/bin/rustfs"); err == nil {
		return true
	}
	return false
}

func upgradeRustFS() error {
	utils.PrintInfo("Downloading RustFS installer...")

	// Download the installation script
	if err := utils.RunCommand("curl", "-O", "https://rustfs.com/install_rustfs.sh"); err != nil {
		return fmt.Errorf("failed to download RustFS installer: %w", err)
	}

	// Make the script executable
	if err := utils.RunCommand("chmod", "+x", "install_rustfs.sh"); err != nil {
		return fmt.Errorf("failed to make installer executable: %w", err)
	}

	// Run the script with sudo and choose option 3 (Upgrade)
	if err := utils.RunCommandShell("echo '3' | sudo bash install_rustfs.sh"); err != nil {
		return fmt.Errorf("failed to upgrade RustFS: %w", err)
	}

	// Clean up the installer script
	if err := os.Remove("install_rustfs.sh"); err != nil {
		utils.PrintWarning(fmt.Sprintf("Failed to clean up installer script: %v", err))
	}

	utils.PrintSuccess("RustFS upgraded successfully!")
	return nil
}
