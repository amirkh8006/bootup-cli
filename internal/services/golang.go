package services

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/amirkh8006/bootup-cli/internal/utils"
)

// GoVersion represents a Go version release
type GoVersion struct {
	Version string `json:"version"`
	Stable  bool   `json:"stable"`
	Files   []struct {
		Filename string `json:"filename"`
		OS       string `json:"os"`
		Arch     string `json:"arch"`
		Version  string `json:"version"`
		Sha256   string `json:"sha256"`
		Size     int64  `json:"size"`
		Kind     string `json:"kind"`
	} `json:"files"`
}

// GoReleaseResponse represents the Go releases API response
type GoReleaseResponse []GoVersion

func InstallGolang() error {
	utils.PrintInfo("Fetching available Go versions...")

	// Fetch available versions
	versions, err := fetchGoVersions()
	if err != nil {
		return fmt.Errorf("failed to fetch Go versions: %w", err)
	}

	// Display versions to user
	selectedVersion, err := displayAndSelectGoVersion(versions)
	if err != nil {
		return fmt.Errorf("failed to select version: %w", err)
	}

	// Install selected version
	return installGoVersion(selectedVersion)
}

func fetchGoVersions() (*GoReleaseResponse, error) {
	resp, err := http.Get("https://go.dev/dl/?mode=json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var versions GoReleaseResponse
	if err := json.Unmarshal(body, &versions); err != nil {
		return nil, err
	}

	return &versions, nil
}

func displayAndSelectGoVersion(versions *GoReleaseResponse) (string, error) {
	// Filter stable versions
	var stableVersions []GoVersion
	for _, version := range *versions {
		if version.Stable {
			stableVersions = append(stableVersions, version)
			if len(stableVersions) >= 10 { // Show last 10 stable versions
				break
			}
		}
	}

	if len(stableVersions) == 0 {
		return "", fmt.Errorf("no stable Go versions found")
	}

	// Display available versions
	fmt.Println("\n📋 Available Go Versions:")
	fmt.Println("==================================================")

	fmt.Printf("🚀 Latest Stable: %s\n\n", stableVersions[0].Version)

	fmt.Println("📦 Stable Versions:")
	for i, version := range stableVersions[:min(5, len(stableVersions))] {
		fmt.Printf("  %d. %s\n", i+1, version.Version)
	}

	fmt.Printf("  %d. Custom version (enter manually)\n\n", len(stableVersions[:min(5, len(stableVersions))])+1)

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

	maxStableChoice := min(5, len(stableVersions))

	// Handle selection
	if choice >= 1 && choice <= maxStableChoice {
		selectedVersion := stableVersions[choice-1].Version
		fmt.Printf("✅ Selected: %s\n", selectedVersion)
		return selectedVersion, nil
	} else if choice == maxStableChoice+1 {
		fmt.Print("Enter Go version (e.g., go1.21.5): ")
		customInput, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		customVersion := strings.TrimSpace(customInput)
		if !strings.HasPrefix(customVersion, "go") {
			customVersion = "go" + customVersion
		}
		fmt.Printf("✅ Selected: %s (Custom)\n", customVersion)
		return customVersion, nil
	} else {
		return "", fmt.Errorf("invalid selection: %d", choice)
	}
}

func installGoVersion(version string) error {
	utils.PrintInfo(fmt.Sprintf("Installing Go %s...", version))

	// Detect architecture
	arch := getGoArch()
	if arch == "" {
		return fmt.Errorf("unsupported architecture: %s", runtime.GOARCH)
	}

	// Construct download URL
	filename := fmt.Sprintf("%s.linux-%s.tar.gz", version, arch)
	downloadURL := fmt.Sprintf("https://go.dev/dl/%s", filename)

	utils.PrintInfo(fmt.Sprintf("Downloading Go from: %s", downloadURL))

	// Create temporary directory
	tmpDir := "/tmp/go-install"
	if err := utils.RunCommand("mkdir", "-p", tmpDir); err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer utils.RunCommand("rm", "-rf", tmpDir)

	// Download Go
	downloadPath := fmt.Sprintf("%s/%s", tmpDir, filename)
	downloadCmd := fmt.Sprintf("wget -O '%s' '%s'", downloadPath, downloadURL)
	if err := utils.RunCommandShell(downloadCmd); err != nil {
		utils.PrintWarning("wget failed, trying curl...")
		downloadCmd = fmt.Sprintf("curl -L -o '%s' '%s'", downloadPath, downloadURL)
		if err := utils.RunCommandShell(downloadCmd); err != nil {
			return fmt.Errorf("failed to download Go: %w", err)
		}
	}

	// Remove existing Go installation
	utils.PrintInfo("Removing any existing Go installation...")
	utils.RunCommand("sudo", "rm", "-rf", "/usr/local/go")

	// Extract Go
	utils.PrintInfo("Extracting Go...")
	if err := utils.RunCommand("sudo", "tar", "-C", "/usr/local", "-xzf", downloadPath); err != nil {
		return fmt.Errorf("failed to extract Go: %w", err)
	}

	// Set up environment variables
	utils.PrintInfo("Setting up environment variables...")

	// Add Go to PATH in various shell configuration files
	shells := []string{".bashrc", ".zshrc", ".profile"}
	homeDir := os.Getenv("HOME")

	goPathExport := "\n# Go Programming Language\nexport PATH=$PATH:/usr/local/go/bin\n"

	for _, shell := range shells {
		shellPath := fmt.Sprintf("%s/%s", homeDir, shell)
		if _, err := os.Stat(shellPath); err == nil {
			// Check if Go PATH is already in the file
			content, err := os.ReadFile(shellPath)
			if err == nil && !strings.Contains(string(content), "/usr/local/go/bin") {
				file, err := os.OpenFile(shellPath, os.O_APPEND|os.O_WRONLY, 0644)
				if err == nil {
					file.WriteString(goPathExport)
					file.Close()
					utils.PrintInfo(fmt.Sprintf("Added Go to PATH in %s", shell))
				}
			}
		}
	}

	// Add Go to system-wide PATH
	systemProfile := "/etc/profile.d/go.sh"
	goSystemPath := "export PATH=$PATH:/usr/local/go/bin\n"

	utils.PrintInfo("Adding Go to system PATH...")
	writeSystemPathCmd := fmt.Sprintf("echo '%s' | sudo tee %s", goSystemPath, systemProfile)
	if err := utils.RunCommandShell(writeSystemPathCmd); err != nil {
		utils.PrintWarning("Failed to add Go to system PATH, but installation succeeded")
	}

	// Set current session PATH for verification
	os.Setenv("PATH", os.Getenv("PATH")+":/usr/local/go/bin")

	// Verify installation
	utils.PrintInfo("Verifying installation...")
	if err := utils.RunCommand("/usr/local/go/bin/go", "version"); err != nil {
		utils.PrintError("Go installation verification failed")
		return err
	}

	utils.PrintSuccess(fmt.Sprintf("Go %s installed successfully! 🎉", version))

	// Display next steps
	fmt.Println("\n📝 Next Steps:")
	fmt.Println("==================================================")
	fmt.Println("1. Restart your terminal or run: source ~/.bashrc (or ~/.zshrc)")
	fmt.Println("2. Verify installation: go version")
	fmt.Println("3. Create your first Go project:")
	fmt.Println("   mkdir hello-world && cd hello-world")
	fmt.Println("   go mod init hello-world")
	fmt.Println("   echo 'package main\n\nimport \"fmt\"\n\nfunc main() {\n    fmt.Println(\"Hello, World!\")\n}' > main.go")
	fmt.Println("   go run main.go")

	// Ask about workspace setup
	fmt.Print("\n🤔 Would you like to create a Go workspace directory? (y/n): ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')

	if strings.ToLower(strings.TrimSpace(input)) == "y" {
		workspaceDir := fmt.Sprintf("%s/go", os.Getenv("HOME"))
		utils.PrintInfo(fmt.Sprintf("Creating Go workspace at %s...", workspaceDir))

		if err := utils.RunCommand("mkdir", "-p", workspaceDir+"/src", workspaceDir+"/bin", workspaceDir+"/pkg"); err != nil {
			utils.PrintError("Failed to create Go workspace, but Go installation was successful")
		} else {
			utils.PrintSuccess("Go workspace created successfully!")
			fmt.Printf("📁 Workspace location: %s\n", workspaceDir)

			// Add GOPATH to shell configs if user wants traditional workspace
			fmt.Print("🤔 Set up traditional GOPATH workspace? (y/n): ")
			gopathInput, _ := reader.ReadString('\n')
			if strings.ToLower(strings.TrimSpace(gopathInput)) == "y" {
				gopathExport := fmt.Sprintf("\nexport GOPATH=%s\nexport PATH=$PATH:$GOPATH/bin\n", workspaceDir)
				for _, shell := range shells {
					shellPath := fmt.Sprintf("%s/%s", os.Getenv("HOME"), shell)
					if _, err := os.Stat(shellPath); err == nil {
						file, err := os.OpenFile(shellPath, os.O_APPEND|os.O_WRONLY, 0644)
						if err == nil {
							file.WriteString(gopathExport)
							file.Close()
						}
					}
				}
				utils.PrintSuccess("GOPATH workspace configured!")
			}
		}
	}

	return nil
}

func getGoArch() string {
	switch runtime.GOARCH {
	case "amd64":
		return "amd64"
	case "arm64":
		return "arm64"
	case "386":
		return "386"
	case "arm":
		return "armv6l"
	default:
		return ""
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
