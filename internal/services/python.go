package services

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/amirkh8006/bootup-cli/internal/utils"
)

func InstallPython() error {
	utils.PrintInfo("Installing Python...")

	// Check if running on supported OS
	if runtime.GOOS != "linux" {
		return fmt.Errorf("Python installation is currently only supported on Linux")
	}

	// For now, assume Ubuntu/Debian (like other services)
	// Can be extended later for other distributions
	return installPythonDebian()
}

func installPythonDebian() error {
	utils.PrintInfo("Installing Python on Ubuntu/Debian...")

	// Update package lists
	if err := utils.RunCommand("sudo", "apt", "update"); err != nil {
		return fmt.Errorf("failed to update package lists: %w", err)
	}

	// Add deadsnakes PPA for additional Python versions (Ubuntu only)
	utils.PrintInfo("Adding deadsnakes PPA for more Python version options...")
	if err := utils.RunCommand("sudo", "apt", "install", "-y", "software-properties-common"); err != nil {
		utils.PrintWarning("Failed to install software-properties-common, continuing with default repositories")
	} else {
		if err := utils.RunCommand("sudo", "add-apt-repository", "-y", "ppa:deadsnakes/ppa"); err != nil {
			utils.PrintWarning("Failed to add deadsnakes PPA, using default repositories")
		} else {
			// Update package lists after adding PPA
			if err := utils.RunCommand("sudo", "apt", "update"); err != nil {
				return fmt.Errorf("failed to update package lists after adding PPA: %w", err)
			}
		}
	}

	// Display available Python versions
	versions := []string{
		"python3.12",
		"python3.11",
		"python3.10",
		"python3.9",
		"python3.8",
	}

	selectedVersion, err := displayAndSelectPythonVersion(versions)
	if err != nil {
		return fmt.Errorf("failed to select version: %w", err)
	}

	// Install Python and essential packages
	packages := []string{
		selectedVersion,
		selectedVersion + "-dev",
		selectedVersion + "-venv",
		selectedVersion + "-distutils",
		"python3-pip",
		"build-essential",
		"libssl-dev",
		"libffi-dev",
		"python3-setuptools",
	}

	utils.PrintInfo(fmt.Sprintf("Installing %s and essential packages...", selectedVersion))
	args := append([]string{"apt", "install", "-y"}, packages...)
	if err := utils.RunCommand("sudo", args...); err != nil {
		return fmt.Errorf("failed to install Python: %w", err)
	}

	// Update alternatives to make the selected version default
	if err := setupPythonAlternatives(selectedVersion); err != nil {
		utils.PrintWarning("Failed to setup Python alternatives: " + err.Error())
	}

	// Install/upgrade pip
	if err := upgradePip(selectedVersion); err != nil {
		utils.PrintWarning("Failed to upgrade pip: " + err.Error())
	}

	// Install common Python packages
	if err := installCommonPackages(); err != nil {
		utils.PrintWarning("Failed to install some common packages: " + err.Error())
	}

	utils.PrintSuccess("Python installation completed successfully!")

	// Display installation info
	return displayPythonInfo(selectedVersion)
}

func displayAndSelectPythonVersion(versions []string) (string, error) {
	if len(versions) == 0 {
		return "", fmt.Errorf("no Python versions available")
	}

	utils.PrintInfo("Available Python versions:")
	for i, version := range versions {
		fmt.Printf("%d. %s\n", i+1, version)
	}

	fmt.Print("\nSelect a version (default: 1): ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}

	input = strings.TrimSpace(input)
	if input == "" {
		input = "1"
	}

	choice, err := strconv.Atoi(input)
	if err != nil || choice < 1 || choice > len(versions) {
		return "", fmt.Errorf("invalid selection: %s", input)
	}

	return versions[choice-1], nil
}

func setupPythonAlternatives(selectedVersion string) error {
	utils.PrintInfo("Setting up Python alternatives...")

	// Setup alternatives for python3
	if err := utils.RunCommand("sudo", "update-alternatives", "--install", "/usr/bin/python3", "python3", "/usr/bin/"+selectedVersion, "1"); err != nil {
		return fmt.Errorf("failed to setup python3 alternative: %w", err)
	}

	// Setup alternatives for python (generic)
	if err := utils.RunCommand("sudo", "update-alternatives", "--install", "/usr/bin/python", "python", "/usr/bin/"+selectedVersion, "1"); err != nil {
		utils.PrintWarning("Failed to setup generic python alternative, this is usually not critical")
	}

	return nil
}

func upgradePip(selectedVersion string) error {
	utils.PrintInfo("Upgrading pip...")

	// First try with the specific Python version
	if err := utils.RunCommand(selectedVersion, "-m", "pip", "install", "--upgrade", "pip"); err != nil {
		// Fallback to python3
		if err := utils.RunCommand("python3", "-m", "pip", "install", "--upgrade", "pip"); err != nil {
			return fmt.Errorf("failed to upgrade pip: %w", err)
		}
	}

	return nil
}

func installCommonPackages() error {
	utils.PrintInfo("Installing common Python packages...")

	packages := []string{
		"virtualenv",
		"wheel",
		"setuptools",
		"requests",
		"numpy",
		"pandas",
		"matplotlib",
		"pytest",
		"black",
		"flake8",
		"pipenv",
	}

	for _, pkg := range packages {
		utils.PrintInfo(fmt.Sprintf("Installing %s...", pkg))
		if err := utils.RunCommand("python3", "-m", "pip", "install", pkg); err != nil {
			utils.PrintWarning(fmt.Sprintf("Failed to install %s: %v", pkg, err))
		}
	}

	return nil
}

func displayPythonInfo(selectedVersion string) error {
	utils.PrintInfo("Python installation summary:")

	// Display Python version
	fmt.Println("Python version:")
	if err := utils.RunCommand("python3", "--version"); err != nil {
		utils.PrintWarning("Failed to get Python version")
	}

	// Display pip version
	fmt.Println("\nPip version:")
	if err := utils.RunCommand("python3", "-m", "pip", "--version"); err != nil {
		utils.PrintWarning("Failed to get pip version")
	}

	// Display installed location
	fmt.Println("\nPython executable location:")
	if err := utils.RunCommand("which", "python3"); err != nil {
		utils.PrintWarning("Failed to get Python location")
	}

	fmt.Println("\n" + strings.Repeat("=", 50))
	utils.PrintSuccess("Python is ready to use!")
	fmt.Println("Tips:")
	fmt.Println("- Create virtual environments: python3 -m venv myenv")
	fmt.Println("- Activate virtual environment: source myenv/bin/activate")
	fmt.Println("- Install packages: pip install package_name")
	fmt.Println("- Use pipenv for project management: pipenv install")
	fmt.Println(strings.Repeat("=", 50))

	return nil
}
