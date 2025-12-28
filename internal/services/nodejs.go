package services

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/amirkh8006/bootup-cli/internal/utils"
)

// NodeVersion represents a Node.js version
type NodeVersion struct {
	Version string      `json:"version"`
	Date    string      `json:"date"`
	Files   []string    `json:"files"`
	Npm     string      `json:"npm"`
	V8      string      `json:"v8"`
	Uv      string      `json:"uv"`
	Zlib    string      `json:"zlib"`
	OpenSSL string      `json:"openssl"`
	Modules string      `json:"modules"`
	LTS     interface{} `json:"lts"`
}

// NodeDistIndex represents the Node.js distribution index
type NodeDistIndex []NodeVersion

func InstallNodeJS() error {
	utils.PrintInfo("Fetching available Node.js versions...")

	// Fetch available versions
	versions, err := fetchNodeVersions()
	if err != nil {
		return fmt.Errorf("failed to fetch Node.js versions: %w", err)
	}

	// Display versions to user
	selectedVersion, err := displayAndSelectVersion(versions)
	if err != nil {
		return fmt.Errorf("failed to select version: %w", err)
	}

	// Install selected version
	return installNodeVersion(selectedVersion)
}

func fetchNodeVersions() (*NodeDistIndex, error) {
	resp, err := http.Get("https://nodejs.org/dist/index.json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var versions NodeDistIndex
	if err := json.Unmarshal(body, &versions); err != nil {
		return nil, err
	}

	return &versions, nil
}

func displayAndSelectVersion(versions *NodeDistIndex) (string, error) {
	// Find current and LTS versions
	var currentVersion string
	var ltsVersions []NodeVersion

	if len(*versions) > 0 {
		currentVersion = (*versions)[0].Version
	}

	// Find LTS versions (last 5 LTS releases)
	ltsCount := 0
	for _, version := range *versions {
		if version.LTS != nil && version.LTS != false {
			ltsVersions = append(ltsVersions, version)
			ltsCount++
			if ltsCount >= 5 {
				break
			}
		}
	}

	// Display available versions
	fmt.Println("\n📋 Available Node.js Versions:")
	fmt.Println("==================================================")

	fmt.Printf("🚀 Current Version: %s\n\n", currentVersion)

	fmt.Println("🔒 LTS (Long Term Support) Versions:")
	for i, version := range ltsVersions {
		ltsName := "Unknown"
		if version.LTS != nil {
			if ltsNameStr, ok := version.LTS.(string); ok {
				ltsName = ltsNameStr
			}
		}
		fmt.Printf("  %d. %s (%s LTS) - npm: %s\n", i+1, version.Version, ltsName, version.Npm)
	}

	fmt.Printf("\n  %d. %s (Current/Latest)\n", len(ltsVersions)+1, currentVersion)
	fmt.Printf("  %d. Custom version (enter manually)\n\n", len(ltsVersions)+2)

	// Get user selection
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Please select a version (enter number): ")
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	choice, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil {
		return "", fmt.Errorf("invalid selection: %s", strings.TrimSpace(input))
	}

	// Handle selection
	if choice >= 1 && choice <= len(ltsVersions) {
		selectedVersion := ltsVersions[choice-1].Version
		fmt.Printf("✅ Selected: %s\n", selectedVersion)
		return selectedVersion, nil
	} else if choice == len(ltsVersions)+1 {
		fmt.Printf("✅ Selected: %s (Current)\n", currentVersion)
		return currentVersion, nil
	} else if choice == len(ltsVersions)+2 {
		fmt.Print("Enter Node.js version (e.g., v18.19.0): ")
		customInput, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		customVersion := strings.TrimSpace(customInput)
		if !strings.HasPrefix(customVersion, "v") {
			customVersion = "v" + customVersion
		}
		fmt.Printf("✅ Selected: %s (Custom)\n", customVersion)
		return customVersion, nil
	} else {
		return "", fmt.Errorf("invalid selection: %d", choice)
	}
}

func installNodeVersion(version string) error {
	// Extract major version number for repository setup
	majorVersion := extractMajorVersion(version)

	utils.PrintInfo(fmt.Sprintf("Installing Node.js %s...", version))

	// Add NodeSource repository
	utils.PrintInfo("Adding NodeSource repository...")
	setupCmd := fmt.Sprintf("curl -fsSL https://deb.nodesource.com/setup_%s.x | sudo -E bash -", majorVersion)
	if err := utils.RunCommandShell(setupCmd); err != nil {
		return fmt.Errorf("failed to add NodeSource repository: %w", err)
	}

	// Update package list
	utils.PrintInfo("Updating package list...")
	if err := utils.RunCommand("sudo", "apt-get", "update", "-y"); err != nil {
		return fmt.Errorf("failed to update package list: %w", err)
	}

	// Install Node.js
	utils.PrintInfo("Installing Node.js and npm...")
	if err := utils.RunCommand("sudo", "apt-get", "install", "-y", "nodejs"); err != nil {
		return fmt.Errorf("failed to install Node.js: %w", err)
	}

	// Verify installation
	utils.PrintInfo("Verifying installation...")
	if err := utils.RunCommand("node", "--version"); err != nil {
		utils.PrintError("Node.js installation verification failed")
		return err
	}

	if err := utils.RunCommand("npm", "--version"); err != nil {
		utils.PrintError("npm installation verification failed")
		return err
	}

	utils.PrintSuccess(fmt.Sprintf("Node.js %s and npm installed successfully! 🎉", version))

	// Optional: Install PM2
	fmt.Print("\n🤔 Would you like to install PM2 (Process Manager)? (y/n): ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')

	if strings.ToLower(strings.TrimSpace(input)) == "y" {
		utils.PrintInfo("Installing PM2...")
		if err := utils.RunCommand("sudo", "npm", "install", "-g", "pm2"); err != nil {
			utils.PrintError("Failed to install PM2, but Node.js installation was successful")
		} else {
			utils.PrintSuccess("PM2 installed successfully!")
		}
	}

	return nil
}

func extractMajorVersion(version string) string {
	// Remove 'v' prefix if present
	version = strings.TrimPrefix(version, "v")

	// Split by '.' and get first part
	parts := strings.Split(version, ".")
	if len(parts) > 0 {
		return parts[0]
	}

	// Fallback to 18 if parsing fails
	return "18"
}
