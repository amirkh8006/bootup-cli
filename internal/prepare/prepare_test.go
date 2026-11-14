package prepare

import (
	"os"
	"strings"
	"testing"
)

func TestFirewallProtectsCurrentSSHConnection(t *testing.T) {
	t.Setenv("SSH_CONNECTION", "192.0.2.12 50000 192.0.2.1 1512")
	o := Defaults()
	if o.SSHPort != "1512" || o.Firewall {
		t.Fatalf("unexpected defaults: %+v", o)
	}
	o.Firewall = true
	for _, tc := range []struct {
		port, source string
		valid        bool
	}{
		{"22", "", false},
		{"1512", "192.0.2.0/24", true},
		{"1512", "192.0.2.12", true},
		{"1512", "198.51.100.0/24", false},
		{"1512", "--anything", false},
	} {
		o.SSHPort, o.SSHSource = tc.port, tc.source
		if err := o.Validate(); (err == nil) != tc.valid {
			t.Errorf("%+v: %v", tc, err)
		}
	}
}

func TestPreparationRejectsInvalidInput(t *testing.T) {
	for _, tc := range []struct{ hostname, timezone, swap string }{
		{"bad hostname", "", "2048"},
		{"-host", "", "2048"},
		{"", "../etc/passwd", "2048"},
		{"", "", "-1"},
	} {
		o := Defaults()
		o.Hostname, o.Timezone, o.SwapMiB = tc.hostname, tc.timezone, tc.swap
		if err := o.Validate(); err == nil {
			t.Errorf("accepted invalid options: %+v", o)
		}
	}
}

func TestCommandEmbedsWorkflowAndCleansUp(t *testing.T) {
	o := Defaults()
	o.Hostname = "production-01"
	cmd, cleanup, err := Command(o)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var scriptPath string
	for _, arg := range cmd.Args {
		if strings.HasSuffix(arg, ".sh") {
			scriptPath = arg
		}
	}
	data, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "configure_swap()") {
		t.Fatal("missing embedded workflow")
	}
	if !strings.Contains(strings.Join(cmd.Args, " "), "--no-firewall") {
		t.Fatal("firewall must be opt-in")
	}
	cleanup()
	if _, err := os.Stat(scriptPath); !os.IsNotExist(err) {
		t.Fatal("temporary script not removed")
	}
}
