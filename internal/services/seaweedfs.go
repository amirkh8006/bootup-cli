package services

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/amirkh8006/bootup-cli/internal/utils"
)

// InstallSeaweedFS installs SeaweedFS distributed file system
func InstallSeaweedFS() error {
	utils.PrintInfo("Installing SeaweedFS...")

	// Otherwise install the binary directly
	utils.PrintInfo("Installing SeaweedFS binary...")
	return installSeaweedFSBinary()
}

// installSeaweedFSBinary installs SeaweedFS binary directly
func installSeaweedFSBinary() error {
	// Create binary directory
	binDir := "/usr/local/bin"

	// Download the latest SeaweedFS binary
	utils.PrintInfo("Downloading SeaweedFS binary...")
	downloadURL := "https://github.com/seaweedfs/seaweedfs/releases/latest/download/linux_amd64.tar.gz"

	if err := utils.RunCommand("wget", "-O", "/tmp/seaweedfs.tar.gz", downloadURL); err != nil {
		return fmt.Errorf("failed to download SeaweedFS: %w", err)
	}

	// Extract the binary
	utils.PrintInfo("Extracting SeaweedFS binary...")
	if err := utils.RunCommand("tar", "-xzf", "/tmp/seaweedfs.tar.gz", "-C", "/tmp/"); err != nil {
		return fmt.Errorf("failed to extract SeaweedFS: %w", err)
	}

	// Install the binary
	utils.PrintInfo("Installing SeaweedFS binary...")
	if err := utils.RunCommand("sudo", "mv", "/tmp/weed", filepath.Join(binDir, "weed")); err != nil {
		return fmt.Errorf("failed to install SeaweedFS binary: %w", err)
	}

	if err := utils.RunCommand("sudo", "chmod", "+x", filepath.Join(binDir, "weed")); err != nil {
		return fmt.Errorf("failed to make SeaweedFS executable: %w", err)
	}

	// Create SeaweedFS data directory
	dataDir := "/var/lib/seaweedfs"
	utils.PrintInfo("Creating SeaweedFS data directory...")
	if err := utils.RunCommand("sudo", "mkdir", "-p", dataDir); err != nil {
		return fmt.Errorf("failed to create SeaweedFS data directory: %w", err)
	}

	// Create systemd service for binary installation
	serviceContent := `[Unit]
Description=SeaweedFS Distributed File System
After=network.target

[Service]
Type=simple
User=root
Group=root
ExecStart=/usr/local/bin/weed server -dir=/var/lib/seaweedfs -s3 -master.volumeSizeLimitMB=1024
Environment=AWS_ACCESS_KEY_ID=admin
Environment=AWS_SECRET_ACCESS_KEY=admin123
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
`

	servicePath := "/tmp/seaweedfs.service"
	if err := os.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
		return fmt.Errorf("failed to write service file: %w", err)
	}

	utils.PrintInfo("Installing systemd service...")
	if err := utils.RunCommand("sudo", "mv", servicePath, "/etc/systemd/system/seaweedfs.service"); err != nil {
		return fmt.Errorf("failed to install service file: %w", err)
	}

	// Reload systemd and enable service
	utils.PrintInfo("Enabling SeaweedFS service...")
	if err := utils.RunCommand("sudo", "systemctl", "daemon-reload"); err != nil {
		return fmt.Errorf("failed to reload systemd: %w", err)
	}

	if err := utils.RunCommand("sudo", "systemctl", "enable", "seaweedfs"); err != nil {
		return fmt.Errorf("failed to enable SeaweedFS service: %w", err)
	}

	// Start SeaweedFS service
	utils.PrintInfo("Starting SeaweedFS service...")
	if err := utils.RunCommand("sudo", "systemctl", "start", "seaweedfs"); err != nil {
		return fmt.Errorf("failed to start SeaweedFS service: %w", err)
	}

	// Cleanup temporary files
	if err := utils.RunCommand("rm", "-f", "/tmp/seaweedfs.tar.gz"); err != nil {
		return fmt.Errorf("failed to clean up temporary files: %w", err)
	}

	utils.PrintSuccess("SeaweedFS installed and started successfully!")
	utils.PrintInfo("SeaweedFS services are available at:")
	utils.PrintInfo("  • Master UI: http://localhost:9333")
	utils.PrintInfo("  • Volume Server: http://localhost:8080")
	utils.PrintInfo("  • Filer UI: http://localhost:8888")
	utils.PrintInfo("  • S3 API: http://localhost:8333")
	utils.PrintInfo("  • WebDAV: http://localhost:7333")
	utils.PrintInfo("S3 Credentials: AccessKey=admin, SecretKey=admin123")

	return nil
}

// isSeaweedFSInstalled checks if SeaweedFS is installed on the system
func isSeaweedFSInstalled() bool {
	// Check if weed binary is available
	if isCommandAvailable("weed") {
		return true
	}

	// Check if systemd service exists
	if _, err := os.Stat("/etc/systemd/system/seaweedfs.service"); err == nil {
		return true
	}

	// Check if Docker compose setup exists
	if _, err := os.Stat("/opt/seaweedfs/docker-compose.yml"); err == nil {
		return true
	}

	return false
}
