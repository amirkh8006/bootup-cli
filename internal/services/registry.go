package services

import "fmt"

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
	"kafka": {
		Name:        "kafka",
		Description: "Distributed streaming platform",
		Category:    "Message Brokers",
		Installer:   InstallKafka,
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
	}
}
