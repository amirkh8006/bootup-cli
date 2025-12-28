package services

import (
	"fmt"
	"github.com/amirkh8006/bootup-cli/internal/utils"
)

func InstallNginx() error {
    utils.PrintInfo("Updating package list...")
    if err := utils.RunCommand("sudo", "apt-get", "update", "-y"); err != nil {
        return fmt.Errorf("failed to update packages: %w", err)
    }

    utils.PrintInfo("Installing Nginx...")
    if err := utils.RunCommand("sudo", "apt-get", "install", "-y", "nginx", "apache2-utils"); err != nil {
        return fmt.Errorf("failed to install Nginx: %w", err)
    }

    utils.PrintSuccess("Nginx installed successfully!")
    return nil
}
