package utils

import (
	"fmt"
	"os"
	"os/exec"
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
