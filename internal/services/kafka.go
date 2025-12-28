package services

import (
	"fmt"
	"os"

	"github.com/amirkh8006/bootup-cli/internal/utils"
)

const (
	kafkaVersion    = "4.1.0"
	kafkaInstallDir = "/opt/kafka"
	kafkaDataDir    = "/var/lib/kafka/data"
	scalaVersion    = "2.13"
)

func InstallKafka() error {
	utils.PrintInfo("Installing Kafka 4.1.0 (KRaft mode)...")

	// Install dependencies
	utils.PrintInfo("Installing Java 17 and dependencies...")
	if err := utils.RunCommand("sudo", "apt-get", "update", "-y"); err != nil {
		return fmt.Errorf("failed to update packages: %w", err)
	}
	if err := utils.RunCommand("sudo", "apt-get", "install", "-y", "openjdk-17-jdk", "wget", "tar", "uuid-runtime"); err != nil {
		return fmt.Errorf("failed to install dependencies: %w", err)
	}

	// Download Kafka
	utils.PrintInfo("Downloading Kafka...")
	kafkaUrl := fmt.Sprintf("https://downloads.apache.org/kafka/%s/kafka_%s-%s.tgz", kafkaVersion, scalaVersion, kafkaVersion)
	if err := utils.RunCommand("wget", kafkaUrl, "-O", "/tmp/kafka.tgz"); err != nil {
		return fmt.Errorf("failed to download Kafka: %w", err)
	}

	// Create installation directory and extract Kafka
	utils.PrintInfo("Extracting Kafka...")
	if err := utils.RunCommand("sudo", "mkdir", "-p", kafkaInstallDir); err != nil {
		return fmt.Errorf("failed to create installation directory: %w", err)
	}
	if err := utils.RunCommand("sudo", "tar", "-xzf", "/tmp/kafka.tgz", "--strip-components=1", "-C", kafkaInstallDir); err != nil {
		return fmt.Errorf("failed to extract Kafka: %w", err)
	}

	// Create data and config directories
	utils.PrintInfo("Setting up directories and permissions...")
	if err := utils.RunCommand("sudo", "mkdir", "-p", kafkaDataDir); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}
	if err := utils.RunCommand("sudo", "mkdir", "-p", kafkaInstallDir+"/config/kraft"); err != nil {
		return fmt.Errorf("failed to create kraft config directory: %w", err)
	}

	// Get current user for ownership
	currentUser := os.Getenv("USER")
	if currentUser == "" {
		currentUser = "ubuntu" // fallback for some environments
	}

	if err := utils.RunCommand("sudo", "chown", currentUser+":"+currentUser, kafkaDataDir); err != nil {
		return fmt.Errorf("failed to set data directory ownership: %w", err)
	}
	if err := utils.RunCommand("sudo", "chmod", "-R", "700", kafkaDataDir); err != nil {
		return fmt.Errorf("failed to set data directory permissions: %w", err)
	}
	if err := utils.RunCommand("sudo", "chown", "-R", currentUser+":"+currentUser, kafkaInstallDir); err != nil {
		return fmt.Errorf("failed to set kafka directory ownership: %w", err)
	}

	// Create KRaft configuration file
	utils.PrintInfo("Creating KRaft configuration...")
	if err := createKraftConfig(); err != nil {
		return fmt.Errorf("failed to create KRaft config: %w", err)
	}

	// Format Kafka storage
	utils.PrintInfo("Formatting Kafka storage for KRaft mode...")
	if err := utils.RunCommandShell("sudo rm -rf /var/lib/kafka/data/*"); err != nil {
		return fmt.Errorf("failed to clean data directory: %w", err)
	}

	// Generate cluster UUID and format storage
	uuidCmd := fmt.Sprintf("%s/bin/kafka-storage.sh format -t $(uuidgen) -c %s/config/kraft/server.properties", kafkaInstallDir, kafkaInstallDir)
	if err := utils.RunCommandShell(uuidCmd); err != nil {
		return fmt.Errorf("failed to format Kafka storage: %w", err)
	}

	// Create systemd service
	utils.PrintInfo("Creating systemd service...")
	if err := createKafkaService(currentUser); err != nil {
		return fmt.Errorf("failed to create systemd service: %w", err)
	}

	// Enable and start service
	utils.PrintInfo("Enabling and starting Kafka service...")
	if err := utils.RunCommand("sudo", "systemctl", "daemon-reload"); err != nil {
		return fmt.Errorf("failed to reload systemd: %w", err)
	}
	if err := utils.RunCommand("sudo", "systemctl", "enable", "kafka"); err != nil {
		return fmt.Errorf("failed to enable Kafka service: %w", err)
	}
	if err := utils.RunCommand("sudo", "systemctl", "start", "kafka"); err != nil {
		return fmt.Errorf("failed to start Kafka service: %w", err)
	}

	// Clean up downloaded file
	utils.RunCommand("rm", "-f", "/tmp/kafka.tgz")

	utils.PrintSuccess("Kafka installation complete!")
	utils.PrintInfo("Kafka is running on localhost:9092")
	utils.PrintInfo("You can check status with: sudo systemctl status kafka")

	return nil
}

func createKraftConfig() error {
	configContent := `# Kafka 4.1.0 KRaft single-node
process.roles=broker,controller
node.id=1
controller.quorum.voters=1@localhost:9093

# Listeners
listeners=PLAINTEXT://:9092,CONTROLLER://:9093
advertised.listeners=PLAINTEXT://localhost:9092
listener.security.protocol.map=PLAINTEXT:PLAINTEXT,CONTROLLER:PLAINTEXT
controller.listener.names=CONTROLLER

# Data
log.dirs=/var/lib/kafka/data
num.partitions=1
offsets.topic.replication.factor=1
transaction.state.log.replication.factor=1
transaction.state.log.min.isr=1
log.retention.hours=168
group.initial.rebalance.delay.ms=0

`

	configPath := kafkaInstallDir + "/config/kraft/server.properties"

	// Create a temporary file with the config content
	tempFile := "/tmp/kafka_config.tmp"
	if err := os.WriteFile(tempFile, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to create temporary config file: %w", err)
	}
	defer os.Remove(tempFile)

	// Copy the config using sudo
	if err := utils.RunCommandShell(fmt.Sprintf("sudo cp %s %s", tempFile, configPath)); err != nil {
		return fmt.Errorf("failed to copy config file: %w", err)
	}

	return nil
}

func createKafkaService(user string) error {
	serviceContent := fmt.Sprintf(`[Unit]
Description=Apache Kafka (KRaft mode)
After=network.target

[Service]
Type=simple
User=%s
Environment="JAVA_HOME=/usr/lib/jvm/java-17-openjdk-amd64"
ExecStart=/opt/kafka/bin/kafka-server-start.sh /opt/kafka/config/kraft/server.properties
ExecStop=/opt/kafka/bin/kafka-server-stop.sh
Restart=on-failure
RestartSec=5
WorkingDirectory=/opt/kafka

[Install]
WantedBy=multi-user.target
`, user)

	servicePath := "/etc/systemd/system/kafka.service"

	// Create a temporary file with the service content
	tempFile := "/tmp/kafka_service.tmp"
	if err := os.WriteFile(tempFile, []byte(serviceContent), 0644); err != nil {
		return fmt.Errorf("failed to create temporary service file: %w", err)
	}
	defer os.Remove(tempFile)

	// Copy the service file using sudo
	if err := utils.RunCommandShell(fmt.Sprintf("sudo cp %s %s", tempFile, servicePath)); err != nil {
		return fmt.Errorf("failed to copy service file: %w", err)
	}

	return nil
}
