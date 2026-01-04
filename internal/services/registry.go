package services

import (
	"fmt"
	"os/exec"
	"strings"
)

// ServiceInfo represents information about a service
type ServiceInfo struct {
	Name        string
	Description string
	Category    string
	Installer   func() error
}

// serviceRegistry contains all available services and their configurations
var serviceRegistry = map[string]ServiceInfo{
	"nginx": {
		Name:        "nginx",
		Description: "High-performance web server",
		Category:    "Web Servers",
		Installer:   InstallNginx,
	},
	"caddy": {
		Name:        "caddy",
		Description: "Modern web server with automatic HTTPS",
		Category:    "Web Servers",
		Installer:   InstallCaddy,
	},
	"postgresql": {
		Name:        "postgresql",
		Description: "Powerful relational database",
		Category:    "Databases",
		Installer:   InstallPostgreSQL,
	},
	"mongodb": {
		Name:        "mongodb",
		Description: "NoSQL document database",
		Category:    "Databases",
		Installer:   InstallMongoDB,
	},
	"redis": {
		Name:        "redis",
		Description: "In-memory data structure store",
		Category:    "Databases",
		Installer:   InstallRedis,
	},
	"elasticsearch": {
		Name:        "elasticsearch",
		Description: "Distributed search and analytics engine",
		Category:    "Databases",
		Installer:   InstallElasticsearch,
	},
	"mysql": {
		Name:        "mysql",
		Description: "Popular open-source relational database",
		Category:    "Databases",
		Installer:   InstallMySQL,
	},
	"clickhouse": {
		Name:        "clickhouse",
		Description: "High-performance columnar database for analytics",
		Category:    "Databases",
		Installer:   InstallClickHouse,
	},
	"nodejs": {
		Name:        "nodejs",
		Description: "JavaScript runtime environment",
		Category:    "Development",
		Installer:   InstallNodeJS,
	},
	"golang": {
		Name:        "golang",
		Description: "Go programming language compiler and tools",
		Category:    "Development",
		Installer:   InstallGolang,
	},
	"php": {
		Name:        "php",
		Description: "PHP programming language and runtime",
		Category:    "Development",
		Installer:   InstallPHP,
	},
	"python": {
		Name:        "python",
		Description: "Python programming language and interpreter",
		Category:    "Development",
		Installer:   InstallPython,
	},
	"kafka": {
		Name:        "kafka",
		Description: "Distributed streaming platform",
		Category:    "Message Brokers",
		Installer:   InstallKafka,
	},
	"rabbitmq": {
		Name:        "rabbitmq",
		Description: "Message broker for distributed applications",
		Category:    "Message Brokers",
		Installer:   InstallRabbitMQ,
	},
	"prometheus": {
		Name:        "prometheus",
		Description: "Monitoring and alerting toolkit",
		Category:    "Monitoring",
		Installer:   InstallPrometheus,
	},
	"grafana": {
		Name:        "grafana",
		Description: "Analytics and monitoring platform",
		Category:    "Monitoring",
		Installer:   InstallGrafana,
	},
	"alertmanager": {
		Name:        "alertmanager",
		Description: "Handles alerts from Prometheus",
		Category:    "Monitoring",
		Installer:   InstallAlertmanager,
	},
	"docker": {
		Name:        "docker",
		Description: "Container platform for building and running applications",
		Category:    "Development",
		Installer:   InstallDocker,
	},
	"rustfs": {
		Name:        "rustfs",
		Description: "High-performance object storage system",
		Category:    "Storage",
		Installer:   InstallRustFS,
	},
	"seaweedfs": {
		Name:        "seaweedfs",
		Description: "Fast distributed storage system for blobs, objects, files, and data lake",
		Category:    "Storage",
		Installer:   InstallSeaweedFS,
	},
	"trivy": {
		Name:        "trivy",
		Description: "Vulnerability scanner for containers and other artifacts",
		Category:    "Security",
		Installer:   InstallTrivy,
	},
	"mongodb_exporter": {
		Name:        "mongodb_exporter",
		Description: "MongoDB metrics exporter for Prometheus",
		Category:    "Prometheus Exporters",
		Installer:   InstallMongoExporter,
	},
	"nginx_exporter": {
		Name:        "nginx_exporter",
		Description: "NGINX metrics exporter for Prometheus",
		Category:    "Prometheus Exporters",
		Installer:   InstallNginxExporter,
	},
	"node_exporter": {
		Name:        "node_exporter",
		Description: "Hardware and OS metrics exporter for Prometheus",
		Category:    "Prometheus Exporters",
		Installer:   InstallNodeExporter,
	},
	"postgres_exporter": {
		Name:        "postgres_exporter",
		Description: "PostgreSQL metrics exporter for Prometheus",
		Category:    "Prometheus Exporters",
		Installer:   InstallPostgresExporter,
	},
	"redis_exporter": {
		Name:        "redis_exporter",
		Description: "Redis metrics exporter for Prometheus",
		Category:    "Prometheus Exporters",
		Installer:   InstallRedisExporter,
	},
}

// GetAllServices returns a list of all available services
func GetAllServices() []ServiceInfo {
	services := make([]ServiceInfo, 0, len(serviceRegistry))
	for _, service := range serviceRegistry {
		services = append(services, service)
	}
	return services
}

// GetServiceNames returns a list of all service names
func GetServiceNames() []string {
	names := make([]string, 0, len(serviceRegistry))
	for name := range serviceRegistry {
		names = append(names, name)
	}
	return names
}

// GetServiceInstaller returns the installer function for a service
func GetServiceInstaller(serviceName string) (func() error, error) {
	service, exists := serviceRegistry[serviceName]
	if !exists {
		return nil, fmt.Errorf("service %s is not supported", serviceName)
	}
	return service.Installer, nil
}

// IsValidService checks if a service name is valid
func IsValidService(serviceName string) bool {
	_, exists := serviceRegistry[serviceName]
	return exists
}

// GetServiceInfo returns information about a specific service
func GetServiceInfo(serviceName string) (ServiceInfo, error) {
	service, exists := serviceRegistry[serviceName]
	if !exists {
		return ServiceInfo{}, fmt.Errorf("service %s is not supported", serviceName)
	}
	return service, nil
}

// GetServicesByCategory returns services organized by category
func GetServicesByCategory() map[string][]ServiceInfo {
	categories := make(map[string][]ServiceInfo)
	for _, service := range serviceRegistry {
		categories[service.Category] = append(categories[service.Category], service)
	}
	return categories
}

// GetCategoryOrder returns the preferred order for displaying categories
func GetCategoryOrder() []string {
	return []string{
		"Web Servers",
		"Databases",
		"Storage",
		"Development",
		"Message Brokers",
		"Monitoring",
		"Prometheus Exporters",
		"Security",
	}
}

// IsServiceInstalled checks if a service is installed on the system
func IsServiceInstalled(serviceName string) bool {
	switch serviceName {
	case "docker":
		return isCommandAvailable("docker")
	case "nginx":
		return isCommandAvailable("nginx")
	case "caddy":
		return isCommandAvailable("caddy")
	case "postgresql":
		return isCommandAvailable("psql") || isCommandAvailable("pg_config")
	case "mongodb":
		return isCommandAvailable("mongod") || isCommandAvailable("mongo")
	case "redis":
		return isCommandAvailable("redis-server") || isCommandAvailable("redis-cli")
	case "elasticsearch":
		return isCommandAvailable("elasticsearch")
	case "mysql":
		return isCommandAvailable("mysql") || isCommandAvailable("mysqld")
	case "clickhouse":
		return isCommandAvailable("clickhouse")
	case "nodejs":
		return isCommandAvailable("node") || isCommandAvailable("nodejs")
	case "golang":
		return isCommandAvailable("go")
	case "php":
		return isCommandAvailable("php")
	case "python":
		return isCommandAvailable("python3") || isCommandAvailable("python")
	case "kafka":
		return isCommandAvailable("kafka-server-start") || isServiceRunning("kafka")
	case "rabbitmq":
		return isCommandAvailable("rabbitmq-server") || isServiceRunning("rabbitmq-server")
	case "prometheus":
		return isCommandAvailable("prometheus") || isServiceRunning("prometheus")
	case "grafana":
		return isCommandAvailable("grafana-server") || isServiceRunning("grafana-server")
	case "alertmanager":
		return isCommandAvailable("alertmanager") || isServiceRunning("alertmanager")
	case "rustfs":
		return isRustFSInstalled()
	case "seaweedfs":
		return isSeaweedFSInstalled()
	case "trivy":
		return isCommandAvailable("trivy")
	case "mongodb_exporter":
		return IsExporterInstalled("mongodb_exporter")
	case "nginx_exporter":
		return IsExporterInstalled("nginx_exporter")
	case "node_exporter":
		return IsExporterInstalled("node_exporter")
	case "postgres_exporter":
		return IsExporterInstalled("postgres_exporter")
	case "redis_exporter":
		return IsExporterInstalled("redis_exporter")
	default:
		return false
	}
}

// isCommandAvailable checks if a command is available in PATH
func isCommandAvailable(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

// isServiceRunning checks if a systemd service is running
func isServiceRunning(serviceName string) bool {
	cmd := exec.Command("systemctl", "is-active", serviceName)
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == "active"
}
