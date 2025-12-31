package services

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/amirkh8006/bootup-cli/internal/utils"
)

const (
	// Installation directory
	installDir = "/usr/local/bin"

	// Exporter Versions
	mongoExporterVersion    = "0.47.1"
	nginxExporterVersion    = "1.5.0"
	nodeExporterVersion     = "1.9.1"
	postgresExporterVersion = "0.17.1"
	redisExporterVersion    = "1.77.0"

	// Default configuration values
	defaultMongoURI       = "mongodb://localhost:27017"
	defaultNginxScrapeURI = "http://127.0.0.1:8080/stub_status"
	defaultPostgresDSN    = "postgresql://postgres:password@localhost:5432/postgres?sslmode=disable"
	defaultRedisAddr      = "redis://localhost:6379"
)

// ExporterConfig holds configuration for each exporter
type ExporterConfig struct {
	MongoURI       string
	NginxScrapeURI string
	PostgresDSN    string
	RedisAddr      string
}

// DefaultExporterConfig returns the default configuration
func DefaultExporterConfig() *ExporterConfig {
	return &ExporterConfig{
		MongoURI:       defaultMongoURI,
		NginxScrapeURI: defaultNginxScrapeURI,
		PostgresDSN:    defaultPostgresDSN,
		RedisAddr:      defaultRedisAddr,
	}
}

// LoadExporterConfig loads configuration from file or returns default
func LoadExporterConfig() *ExporterConfig {
	config := DefaultExporterConfig()

	// Try to load from user config
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return config
	}

	configPath := filepath.Join(homeDir, ".config", "bootup", "exporters.conf")
	file, err := os.Open(configPath)
	if err != nil {
		return config // Return default if config file doesn't exist
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), "\"'")

		switch key {
		case "MONGO_URI":
			config.MongoURI = value
		case "NGINX_SCRAPE_URI":
			config.NginxScrapeURI = value
		case "POSTGRES_DSN":
			config.PostgresDSN = value
		case "REDIS_ADDR":
			config.RedisAddr = value
		}
	}

	return config
}

// InstallMongoExporter installs MongoDB Exporter
func InstallMongoExporter() error {
	return installMongoExporter(LoadExporterConfig())
}

// InstallNginxExporter installs NGINX Exporter
func InstallNginxExporter() error {
	return installNginxExporter(LoadExporterConfig())
}

// InstallNodeExporter installs Node Exporter
func InstallNodeExporter() error {
	return installNodeExporter()
}

// InstallPostgresExporter installs Postgres Exporter
func InstallPostgresExporter() error {
	return installPostgresExporter(LoadExporterConfig())
}

// InstallRedisExporter installs Redis Exporter
func InstallRedisExporter() error {
	return installRedisExporter(LoadExporterConfig())
}

// installMongoExporter installs MongoDB Exporter with configuration
func installMongoExporter(config *ExporterConfig) error {
	fmt.Println("🍃 Installing MongoDB Exporter v" + mongoExporterVersion + "...")

	// Download URL
	url := fmt.Sprintf("https://github.com/percona/mongodb_exporter/releases/download/v%s/mongodb_exporter-%s.linux-amd64.tar.gz",
		mongoExporterVersion, mongoExporterVersion)

	// Download and extract
	tmpDir, err := os.MkdirTemp("", "mongodb_exporter")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	filename := filepath.Join(tmpDir, "mongodb_exporter.tar.gz")
	if err := utils.DownloadFile(url, filename); err != nil {
		return fmt.Errorf("failed to download MongoDB Exporter: %v", err)
	}

	extractDir := filepath.Join(tmpDir, "extract")
	if err := utils.ExtractTarGz(filename, extractDir); err != nil {
		return fmt.Errorf("failed to extract MongoDB Exporter: %v", err)
	}

	// Find and move the binary
	binaryPath := filepath.Join(extractDir, fmt.Sprintf("mongodb_exporter-%s.linux-amd64", mongoExporterVersion), "mongodb_exporter")
	if err := utils.MoveBinaryToSystem(binaryPath, "mongodb_exporter"); err != nil {
		return fmt.Errorf("failed to install MongoDB Exporter binary: %v", err)
	}

	// Create systemd service
	serviceContent := fmt.Sprintf(`[Unit]
Description=MongoDB Exporter
After=network.target

[Service]
ExecStart=%s/mongodb_exporter --mongodb.uri="%s"
Restart=always
User=nobody
Group=nobody

[Install]
WantedBy=multi-user.target
`, installDir, config.MongoURI)

	if err := utils.CreateSystemdService("mongodb_exporter", serviceContent); err != nil {
		return fmt.Errorf("failed to create MongoDB Exporter service: %v", err)
	}

	// Enable and start service
	if err := utils.EnableAndStartService("mongodb_exporter"); err != nil {
		return fmt.Errorf("failed to start MongoDB Exporter: %v", err)
	}

	fmt.Println("✅ MongoDB Exporter installed and started successfully!")
	return nil
}

// installNginxExporter installs NGINX Exporter with configuration
func installNginxExporter(config *ExporterConfig) error {
	fmt.Println("🌐 Installing NGINX Exporter v" + nginxExporterVersion + "...")

	// Download URL
	url := fmt.Sprintf("https://github.com/nginxinc/nginx-prometheus-exporter/releases/download/v%s/nginx-prometheus-exporter_%s_linux_amd64.tar.gz",
		nginxExporterVersion, nginxExporterVersion)

	// Download and extract
	tmpDir, err := os.MkdirTemp("", "nginx_exporter")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	filename := filepath.Join(tmpDir, "nginx_exporter.tar.gz")
	if err := utils.DownloadFile(url, filename); err != nil {
		return fmt.Errorf("failed to download NGINX Exporter: %v", err)
	}

	extractDir := filepath.Join(tmpDir, "extract")
	if err := utils.ExtractTarGz(filename, extractDir); err != nil {
		return fmt.Errorf("failed to extract NGINX Exporter: %v", err)
	}

	// Find and move the binary
	binaryPath := filepath.Join(extractDir, "nginx-prometheus-exporter")
	if err := utils.MoveBinaryToSystem(binaryPath, "nginx-prometheus-exporter"); err != nil {
		return fmt.Errorf("failed to install NGINX Exporter binary: %v", err)
	}

	// Create systemd service
	serviceContent := fmt.Sprintf(`[Unit]
Description=NGINX Exporter
After=network.target

[Service]
ExecStart=%s/nginx-prometheus-exporter -nginx.scrape-uri %s
Restart=always
User=nobody
Group=nobody

[Install]
WantedBy=multi-user.target
`, installDir, config.NginxScrapeURI)

	if err := utils.CreateSystemdService("nginx_exporter", serviceContent); err != nil {
		return fmt.Errorf("failed to create NGINX Exporter service: %v", err)
	}

	// Enable and start service
	if err := utils.EnableAndStartService("nginx_exporter"); err != nil {
		return fmt.Errorf("failed to start NGINX Exporter: %v", err)
	}

	fmt.Println("✅ NGINX Exporter installed and started successfully!")
	return nil
}

// installNodeExporter installs Node Exporter
func installNodeExporter() error {
	fmt.Println("🖥️ Installing Node Exporter v" + nodeExporterVersion + "...")

	// Download URL
	url := fmt.Sprintf("https://github.com/prometheus/node_exporter/releases/download/v%s/node_exporter-%s.linux-amd64.tar.gz",
		nodeExporterVersion, nodeExporterVersion)

	// Download and extract
	tmpDir, err := os.MkdirTemp("", "node_exporter")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	filename := filepath.Join(tmpDir, "node_exporter.tar.gz")
	if err := utils.DownloadFile(url, filename); err != nil {
		return fmt.Errorf("failed to download Node Exporter: %v", err)
	}

	extractDir := filepath.Join(tmpDir, "extract")
	if err := utils.ExtractTarGz(filename, extractDir); err != nil {
		return fmt.Errorf("failed to extract Node Exporter: %v", err)
	}

	// Find and move the binary
	binaryPath := filepath.Join(extractDir, fmt.Sprintf("node_exporter-%s.linux-amd64", nodeExporterVersion), "node_exporter")
	if err := utils.MoveBinaryToSystem(binaryPath, "node_exporter"); err != nil {
		return fmt.Errorf("failed to install Node Exporter binary: %v", err)
	}

	// Create systemd service
	serviceContent := fmt.Sprintf(`[Unit]
Description=Node Exporter
After=network.target

[Service]
ExecStart=%s/node_exporter
Restart=always
User=nobody
Group=nobody

[Install]
WantedBy=multi-user.target
`, installDir)

	if err := utils.CreateSystemdService("node_exporter", serviceContent); err != nil {
		return fmt.Errorf("failed to create Node Exporter service: %v", err)
	}

	// Enable and start service
	if err := utils.EnableAndStartService("node_exporter"); err != nil {
		return fmt.Errorf("failed to start Node Exporter: %v", err)
	}

	fmt.Println("✅ Node Exporter installed and started successfully!")
	return nil
}

// installPostgresExporter installs Postgres Exporter with configuration
func installPostgresExporter(config *ExporterConfig) error {
	fmt.Println("🐘 Installing Postgres Exporter v" + postgresExporterVersion + "...")

	// Download URL
	url := fmt.Sprintf("https://github.com/prometheus-community/postgres_exporter/releases/download/v%s/postgres_exporter-%s.linux-amd64.tar.gz",
		postgresExporterVersion, postgresExporterVersion)

	// Download and extract
	tmpDir, err := os.MkdirTemp("", "postgres_exporter")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	filename := filepath.Join(tmpDir, "postgres_exporter.tar.gz")
	if err := utils.DownloadFile(url, filename); err != nil {
		return fmt.Errorf("failed to download Postgres Exporter: %v", err)
	}

	extractDir := filepath.Join(tmpDir, "extract")
	if err := utils.ExtractTarGz(filename, extractDir); err != nil {
		return fmt.Errorf("failed to extract Postgres Exporter: %v", err)
	}

	// Find and move the binary
	binaryPath := filepath.Join(extractDir, fmt.Sprintf("postgres_exporter-%s.linux-amd64", postgresExporterVersion), "postgres_exporter")
	if err := utils.MoveBinaryToSystem(binaryPath, "postgres_exporter"); err != nil {
		return fmt.Errorf("failed to install Postgres Exporter binary: %v", err)
	}

	// Create systemd service
	serviceContent := fmt.Sprintf(`[Unit]
Description=Postgres Exporter
After=network.target

[Service]
ExecStart=%s/postgres_exporter
Environment="DATA_SOURCE_NAME=%s"
Restart=always
User=nobody
Group=nobody

[Install]
WantedBy=multi-user.target
`, installDir, config.PostgresDSN)

	if err := utils.CreateSystemdService("postgres_exporter", serviceContent); err != nil {
		return fmt.Errorf("failed to create Postgres Exporter service: %v", err)
	}

	// Enable and start service
	if err := utils.EnableAndStartService("postgres_exporter"); err != nil {
		return fmt.Errorf("failed to start Postgres Exporter: %v", err)
	}

	fmt.Println("✅ Postgres Exporter installed and started successfully!")
	return nil
}

// installRedisExporter installs Redis Exporter with configuration
func installRedisExporter(config *ExporterConfig) error {
	fmt.Println("🟥 Installing Redis Exporter v" + redisExporterVersion + "...")

	// Download URL
	url := fmt.Sprintf("https://github.com/oliver006/redis_exporter/releases/download/v%s/redis_exporter-v%s.linux-amd64.tar.gz",
		redisExporterVersion, redisExporterVersion)

	// Download and extract
	tmpDir, err := os.MkdirTemp("", "redis_exporter")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	filename := filepath.Join(tmpDir, "redis_exporter.tar.gz")
	if err := utils.DownloadFile(url, filename); err != nil {
		return fmt.Errorf("failed to download Redis Exporter: %v", err)
	}

	extractDir := filepath.Join(tmpDir, "extract")
	if err := utils.ExtractTarGz(filename, extractDir); err != nil {
		return fmt.Errorf("failed to extract Redis Exporter: %v", err)
	}

	// Find and move the binary
	binaryPath := filepath.Join(extractDir, fmt.Sprintf("redis_exporter-v%s.linux-amd64", redisExporterVersion), "redis_exporter")
	if err := utils.MoveBinaryToSystem(binaryPath, "redis_exporter"); err != nil {
		return fmt.Errorf("failed to install Redis Exporter binary: %v", err)
	}

	// Create systemd service
	serviceContent := fmt.Sprintf(`[Unit]
Description=Redis Exporter
After=network.target

[Service]
ExecStart=%s/redis_exporter --redis.addr=%s
Restart=always
User=nobody
Group=nobody

[Install]
WantedBy=multi-user.target
`, installDir, config.RedisAddr)

	if err := utils.CreateSystemdService("redis_exporter", serviceContent); err != nil {
		return fmt.Errorf("failed to create Redis Exporter service: %v", err)
	}

	// Enable and start service
	if err := utils.EnableAndStartService("redis_exporter"); err != nil {
		return fmt.Errorf("failed to start Redis Exporter: %v", err)
	}

	fmt.Println("✅ Redis Exporter installed and started successfully!")
	return nil
}

// IsExporterInstalled checks if an exporter is installed
func IsExporterInstalled(exporterName string) bool {
	switch exporterName {
	case "mongodb_exporter":
		return isCommandAvailable("mongodb_exporter")
	case "nginx_exporter":
		return isCommandAvailable("nginx-prometheus-exporter")
	case "node_exporter":
		return isCommandAvailable("node_exporter")
	case "postgres_exporter":
		return isCommandAvailable("postgres_exporter")
	case "redis_exporter":
		return isCommandAvailable("redis_exporter")
	default:
		return false
	}
}
