package tunmode

import (
	"fmt"
	"os/exec"
	"strings"
)

func GetDefaultInterfaceAndIP() (name string, ip string, err error) {
	out, err := exec.Command("ip", "route", "get", "142.250.186.78").Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to get default route: %w", err)
	}
	// 142.250.186.78 via 192.168.1.1 dev wlp3s0 src 192.168.1.2 uid 1000
	output := string(out)
	fields := strings.Fields(output)
	var iface, gatewayIP string
	for i := range fields {
		if fields[i] == "dev" && i+1 < len(fields) {
			iface = fields[i+1]
		}
		if fields[i] == "via" && i+1 < len(fields) {
			gatewayIP = fields[i+1]
		}
	}
	if iface == "" || gatewayIP == "" {
		return "", "", fmt.Errorf("could not parse interface or gateway IP from output: %s", output)
	}
	return iface, gatewayIP, nil
}
