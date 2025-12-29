package services

import (
	"fmt"

	"github.com/amirkh8006/bootup-cli/internal/utils"
)

func InstallRabbitMQ() error {
	utils.PrintInfo("Installing RabbitMQ...")

	// Update package lists
	utils.PrintInfo("Updating package lists...")
	if err := utils.RunCommand("sudo", "apt-get", "update", "-y"); err != nil {
		return fmt.Errorf("failed to update package lists: %w", err)
	}

	// Install dependencies
	utils.PrintInfo("Installing dependencies...")
	if err := utils.RunCommand("sudo", "apt-get", "install", "-y", "curl", "gnupg", "apt-transport-https"); err != nil {
		return fmt.Errorf("failed to install dependencies: %w", err)
	}

	// Add RabbitMQ signing key
	utils.PrintInfo("Adding RabbitMQ signing key...")
	if err := utils.RunCommandShell("curl -1sLf 'https://keys.openpgp.org/vks/v1/by-fingerprint/0A9AF2115F4687BD29803A206B73A36E6026DFCA' | sudo gpg --dearmor | sudo tee /usr/share/keyrings/com.rabbitmq.team.gpg > /dev/null"); err != nil {
		return fmt.Errorf("failed to add RabbitMQ signing key: %w", err)
	}

	// Add Erlang Solutions signing key
	utils.PrintInfo("Adding Erlang Solutions signing key...")
	if err := utils.RunCommandShell("curl -1sLf 'https://github.com/rabbitmq/signing-keys/releases/download/3.0/cloudsmith.rabbitmq-erlang.E495BB49CC4BBE5B.key' | sudo gpg --dearmor | sudo tee /usr/share/keyrings/io.cloudsmith.rabbitmq.E495BB49CC4BBE5B.gpg > /dev/null"); err != nil {
		return fmt.Errorf("failed to add Erlang Solutions signing key: %w", err)
	}

	// Add RabbitMQ server signing key
	utils.PrintInfo("Adding RabbitMQ server signing key...")
	if err := utils.RunCommandShell("curl -1sLf 'https://github.com/rabbitmq/signing-keys/releases/download/3.0/cloudsmith.rabbitmq-server.9F4587F226208342.key' | sudo gpg --dearmor | sudo tee /usr/share/keyrings/io.cloudsmith.rabbitmq.9F4587F226208342.gpg > /dev/null"); err != nil {
		return fmt.Errorf("failed to add RabbitMQ server signing key: %w", err)
	}

	// Add RabbitMQ repository
	utils.PrintInfo("Adding RabbitMQ repository...")
	repoConfig := `## Provides modern Erlang/OTP releases
deb [arch=amd64 signed-by=/usr/share/keyrings/io.cloudsmith.rabbitmq.E495BB49CC4BBE5B.gpg] https://dl.cloudsmith.io/public/rabbitmq/rabbitmq-erlang/deb/ubuntu jammy main
deb-src [arch=amd64 signed-by=/usr/share/keyrings/io.cloudsmith.rabbitmq.E495BB49CC4BBE5B.gpg] https://dl.cloudsmith.io/public/rabbitmq/rabbitmq-erlang/deb/ubuntu jammy main

## Provides RabbitMQ
deb [arch=amd64 signed-by=/usr/share/keyrings/io.cloudsmith.rabbitmq.9F4587F226208342.gpg] https://dl.cloudsmith.io/public/rabbitmq/rabbitmq-server/deb/ubuntu jammy main
deb-src [arch=amd64 signed-by=/usr/share/keyrings/io.cloudsmith.rabbitmq.9F4587F226208342.gpg] https://dl.cloudsmith.io/public/rabbitmq/rabbitmq-server/deb/ubuntu jammy main`

	if err := utils.WriteToFile("/tmp/rabbitmq.list", repoConfig); err != nil {
		return fmt.Errorf("failed to create repository configuration: %w", err)
	}

	if err := utils.RunCommand("sudo", "mv", "/tmp/rabbitmq.list", "/etc/apt/sources.list.d/rabbitmq.list"); err != nil {
		return fmt.Errorf("failed to add repository configuration: %w", err)
	}

	// Update package lists with new repository
	utils.PrintInfo("Updating package lists with RabbitMQ repository...")
	if err := utils.RunCommand("sudo", "apt-get", "update", "-y"); err != nil {
		return fmt.Errorf("failed to update package lists: %w", err)
	}

	// Install Erlang packages
	utils.PrintInfo("Installing Erlang packages...")
	erlangPackages := []string{
		"erlang-base",
		"erlang-asn1", "erlang-crypto", "erlang-eldap", "erlang-ftp", "erlang-inets",
		"erlang-mnesia", "erlang-os-mon", "erlang-parsetools", "erlang-public-key",
		"erlang-runtime-tools", "erlang-snmp", "erlang-ssl",
		"erlang-syntax-tools", "erlang-tftp", "erlang-tools", "erlang-xmerl",
	}

	args := append([]string{"apt-get", "install", "-y"}, erlangPackages...)
	if err := utils.RunCommand("sudo", args...); err != nil {
		return fmt.Errorf("failed to install Erlang packages: %w", err)
	}

	// Install RabbitMQ server
	utils.PrintInfo("Installing RabbitMQ server...")
	if err := utils.RunCommand("sudo", "apt-get", "install", "-y", "rabbitmq-server"); err != nil {
		return fmt.Errorf("failed to install RabbitMQ server: %w", err)
	}

	// Enable RabbitMQ service
	utils.PrintInfo("Enabling RabbitMQ service...")
	if err := utils.RunCommand("sudo", "systemctl", "enable", "rabbitmq-server"); err != nil {
		return fmt.Errorf("failed to enable RabbitMQ service: %w", err)
	}

	// Start RabbitMQ service
	utils.PrintInfo("Starting RabbitMQ service...")
	if err := utils.RunCommand("sudo", "systemctl", "start", "rabbitmq-server"); err != nil {
		return fmt.Errorf("failed to start RabbitMQ service: %w", err)
	}

	// Enable RabbitMQ Management Plugin
	utils.PrintInfo("Enabling RabbitMQ Management Plugin...")
	if err := utils.RunCommand("sudo", "rabbitmq-plugins", "enable", "rabbitmq_management"); err != nil {
		return fmt.Errorf("failed to enable management plugin: %w", err)
	}

	// Create admin user
	utils.PrintInfo("Creating admin user...")
	if err := utils.RunCommand("sudo", "rabbitmqctl", "add_user", "admin", "admin"); err != nil {
		utils.PrintWarning("Admin user may already exist")
	}

	if err := utils.RunCommand("sudo", "rabbitmqctl", "set_user_tags", "admin", "administrator"); err != nil {
		return fmt.Errorf("failed to set admin user tags: %w", err)
	}

	if err := utils.RunCommand("sudo", "rabbitmqctl", "set_permissions", "-p", "/", "admin", ".*", ".*", ".*"); err != nil {
		return fmt.Errorf("failed to set admin permissions: %w", err)
	}

	// Restart RabbitMQ to apply changes
	utils.PrintInfo("Restarting RabbitMQ to apply configuration...")
	if err := utils.RunCommand("sudo", "systemctl", "restart", "rabbitmq-server"); err != nil {
		return fmt.Errorf("failed to restart RabbitMQ service: %w", err)
	}

	utils.PrintSuccess("RabbitMQ installed and started successfully!")
	utils.PrintInfo("Management UI is available at http://localhost:15672")
	utils.PrintInfo("Default credentials: admin/admin")
	utils.PrintInfo("AMQP port: 5672")
	utils.PrintInfo("Management port: 15672")

	return nil
}
