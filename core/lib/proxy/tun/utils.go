package tunmode

import (
	"fmt"
	"os/exec"
	"strings"
)

// Returns (name , ip, err)
func GetDefaultInterfaceAndIP() (string, string, error) {
	// Run `ip route get 142.250.186.78` to determine default route
	out, err := exec.Command("ip", "route", "get", "142.250.186.78").Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to get default route: %w", err)
	}

	// Example output:
	// 142.250.186.78 via 192.168.1.1 dev wlp3s0 src 192.168.1.2 uid 1000
	output := string(out)
	fields := strings.Fields(output)

	var iface, srcIP string
	for i := range fields {
		if fields[i] == "dev" && i+1 < len(fields) {
			iface = fields[i+1]
		}
		if fields[i] == "src" && i+1 < len(fields) {
			srcIP = fields[i+1]
		}
	}

	if iface == "" || srcIP == "" {
		return "", "", fmt.Errorf("could not parse interface or IP from output: %s", output)
	}

	return iface, srcIP, nil
}
