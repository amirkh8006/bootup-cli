// Package prepare provides the server preparation workflow used by the TUI.
package prepare

import (
	_ "embed"
	"fmt"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

//go:embed prepare.sh
var script string

type Options struct {
	Hostname, Timezone, SwapMiB, SSHPort, SSHSource string
	Firewall, Upgrade                               bool
}

func Defaults() Options {
	port := "22"
	if fields := strings.Fields(os.Getenv("SSH_CONNECTION")); len(fields) == 4 {
		port = fields[3]
	}
	return Options{SwapMiB: "2048", SSHPort: port}
}

func (o Options) Validate() error {
	if o.Hostname != "" && (len(o.Hostname) > 63 || !regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?$`).MatchString(o.Hostname)) {
		return fmt.Errorf("hostname must be a DNS label of at most 63 characters")
	}
	if !regexp.MustCompile(`^[0-9]{1,7}$`).MatchString(o.SwapMiB) {
		return fmt.Errorf("swap size must be a nonnegative integer in MiB (0 disables)")
	}
	if o.Timezone != "" {
		if strings.HasPrefix(o.Timezone, "/") || strings.Contains(o.Timezone, "..") {
			return fmt.Errorf("invalid timezone")
		}
		if info, err := os.Stat("/usr/share/zoneinfo/" + o.Timezone); err != nil || !info.Mode().IsRegular() {
			return fmt.Errorf("unknown timezone: %s", o.Timezone)
		}
	}
	if !o.Firewall {
		return nil
	}
	port, err := strconv.Atoi(o.SSHPort)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("SSH port must be between 1 and 65535")
	}
	var sourceNet *net.IPNet
	if o.SSHSource != "" {
		ip := net.ParseIP(o.SSHSource)
		if ip == nil {
			_, sourceNet, err = net.ParseCIDR(o.SSHSource)
			if err != nil {
				return fmt.Errorf("SSH source must be an IP address or CIDR")
			}
		} else {
			bits := 128
			if ip.To4() != nil {
				bits = 32
			}
			sourceNet = &net.IPNet{IP: ip, Mask: net.CIDRMask(bits, bits)}
		}
	}
	if fields := strings.Fields(os.Getenv("SSH_CONNECTION")); len(fields) == 4 {
		if fields[3] != o.SSHPort {
			return fmt.Errorf("SSH port must match the current connection (%s)", fields[3])
		}
		if sourceNet != nil && !sourceNet.Contains(net.ParseIP(fields[0])) {
			return fmt.Errorf("SSH source must include your current client (%s)", fields[0])
		}
	}
	return nil
}

// Command stages the embedded workflow so sudo can read it without consuming
// terminal input. The caller must clean up after execution.
func Command(o Options) (*exec.Cmd, func(), error) {
	if err := o.Validate(); err != nil {
		return nil, nil, err
	}
	f, err := os.CreateTemp("", "bootup-prepare-*.sh")
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() { os.Remove(f.Name()) }
	if _, err = f.WriteString(script); err != nil {
		f.Close()
		cleanup()
		return nil, nil, err
	}
	if err = f.Close(); err != nil {
		cleanup()
		return nil, nil, err
	}
	args := []string{f.Name(), "--swap-mib", o.SwapMiB, "--ssh-port", o.SSHPort}
	if o.Hostname != "" {
		args = append(args, "--hostname", o.Hostname)
	}
	if o.Timezone != "" {
		args = append(args, "--timezone", o.Timezone)
	}
	if o.SSHSource != "" {
		args = append(args, "--ssh-source", o.SSHSource)
	}
	if o.Firewall {
		args = append(args, "--firewall")
	} else {
		args = append(args, "--no-firewall")
	}
	if o.Upgrade {
		args = append(args, "--upgrade")
	}
	if os.Geteuid() != 0 {
		return exec.Command("sudo", append([]string{"--", "bash"}, args...)...), cleanup, nil
	}
	return exec.Command("bash", args...), cleanup, nil
}
