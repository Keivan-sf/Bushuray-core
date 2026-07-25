package tunscripts

import (
	"fmt"
	"strings"
)

func GenerateBxrayUser() error {
	_, err := runScriptWithSh(`grep -qw bxray_tproxy /etc/passwd || echo "bxray_tproxy:x:0:23333:::" >> /etc/passwd`)
	return err
}

func GenerateRoutingTable() error {
	script := `
ip rule add fwmark 127 table 102
ip route add local 0.0.0.0/0 dev lo table 102`
	_, err := runScriptWithSh(script)

	return err
}

func BypassLocalAndLan() error {
	local_subnets, err := runScriptWithSh(`
		ip address | grep -w inet | awk '{print $2}'
	`)
	if err != nil {
		return err
	}
	subnets := extractSubnets(local_subnets)

	script := []string{"iptables -t mangle -N BXRAY"}
	for _, subnet := range subnets {
		script = append(script, fmt.Sprintf(`iptables -t mangle -A BXRAY -d %s -j RETURN`, subnet))
	}
	script = append(script, `iptables -t mangle -A BXRAY -d 224.0.0.0/3 -j RETURN`)

	_, err = runScriptWithSh(strings.Join(script, "\n"))
	return err
}

func MarkAndForwardPackets() error {
	script := `
iptables -t mangle -A BXRAY -p tcp -j TPROXY --on-port 13345 --tproxy-mark 127
iptables -t mangle -A BXRAY -p udp -j TPROXY --on-port 13345 --tproxy-mark 127
	`
	_, err := runScriptWithSh(script)
	return err
}

func ApplyMainRules() error {
	script := `iptables -t mangle -A PREROUTING -j BXRAY`
	_, err := runScriptWithSh(script)
	return err
}

func ProxyGateway() error {
	local_subnets, err := runScriptWithSh(`
		ip address | grep -w inet | awk '{print $2}'
	`)
	if err != nil {
		return err
	}
	subnets := extractSubnets(local_subnets)

	script := []string{
		"iptables -t mangle -N BXRAY_MASK",
		"iptables -t mangle -A BXRAY_MASK -m owner --gid-owner 24333 -j RETURN",
	}

	for _, subnet := range subnets {
		script = append(script, fmt.Sprintf(`iptables -t mangle -A BXRAY_MASK -d %s -j RETURN`, subnet))
	}
	script = append(script, `iptables -t mangle -A BXRAY_MASK -d 224.0.0.0/3 -j RETURN`)

	script = append(script, "iptables -t mangle -A BXRAY_MASK -j MARK --set-mark 127")
	script = append(script, "iptables -t mangle -A OUTPUT -p tcp -j BXRAY_MASK")
	script = append(script, "iptables -t mangle -A OUTPUT -p udp -j BXRAY_MASK")

	_, err = runScriptWithSh(strings.Join(script, "\n"))
	return err
}

func BypassDns() error {
	script := `iptables -t mangle -I BXRAY_MASK 1 -p udp --dport 53 -j RETURN`
	_, err := runScriptWithSh(script)
	return err
}

func CleanUp() error {
	script := `
iptables -t mangle -D PREROUTING -j BXRAY
iptables -t mangle -D OUTPUT -p tcp -j BXRAY_MASK
iptables -t mangle -D OUTPUT -p udp -j BXRAY_MASK

iptables -t mangle -F BXRAY
iptables -t mangle -F BXRAY_MASK

iptables -t mangle -X BXRAY
iptables -t mangle -X BXRAY_MASK

ip route del local 0.0.0.0/0 dev lo table 102
ip rule del fwmark 127 table 102
`
	_, err := runScriptWithSh(script)
	return err
}
