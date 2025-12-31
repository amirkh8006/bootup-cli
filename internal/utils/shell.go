package utils

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

func RunCommand(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func RunCommandShell(command string) error {
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func PrintInfo(msg string) {
	fmt.Printf("ℹ️  %s\n", msg)
}

func PrintSuccess(msg string) {
	fmt.Printf("✅ %s\n", msg)
}

func PrintError(msg string) {
	fmt.Printf("❌ %s\n", msg)
}

func PrintWarning(msg string) {
	fmt.Printf("⚠️  %s\n", msg)
}

func WriteToFile(filename, content string) error {
	return os.WriteFile(filename, []byte(content), 0644)
}

// DownloadFile downloads a file from a URL to a local path
func DownloadFile(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// ExtractTarGz extracts a tar.gz file to a destination directory
func ExtractTarGz(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(dest, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}

			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()
		}
	}

	return nil
}

// MoveBinaryToSystem moves a binary to the system path with sudo
func MoveBinaryToSystem(srcPath, binaryName string) error {
	destPath := filepath.Join("/usr/local/bin", binaryName)
	return RunCommand("sudo", "mv", srcPath, destPath)
}

// CreateSystemdService creates a systemd service file
func CreateSystemdService(serviceName, serviceContent string) error {
	servicePath := fmt.Sprintf("/etc/systemd/system/%s.service", serviceName)

	// Write service content to a temp file first
	tmpFile, err := os.CreateTemp("", serviceName+"*.service")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(serviceContent); err != nil {
		tmpFile.Close()
		return err
	}
	tmpFile.Close()

	// Move to systemd directory with sudo
	if err := RunCommand("sudo", "mv", tmpFile.Name(), servicePath); err != nil {
		return err
	}

	// Reload systemd daemon
	return RunCommand("sudo", "systemctl", "daemon-reload")
}

// EnableAndStartService enables and starts a systemd service
func EnableAndStartService(serviceName string) error {
	if err := RunCommand("sudo", "systemctl", "enable", serviceName); err != nil {
		return err
	}
	return RunCommand("sudo", "systemctl", "start", serviceName)
}
