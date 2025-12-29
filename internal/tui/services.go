package tui

import (
	"github.com/amirkh8006/bootup-cli/internal/services"
)

// GetServiceInstaller returns the appropriate installation function for TUI mode
func GetServiceInstaller(serviceName string) func() error {
	installer, err := services.GetServiceInstaller(serviceName)
	if err != nil {
		return func() error {
			return err
		}
	}
	return installer
}
