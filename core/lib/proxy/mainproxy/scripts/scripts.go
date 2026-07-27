package tunscripts

import (
	"fmt"
	"strings"
)

func GenerateBxrayUser() error {
	_, err := runScriptWithSh(`
set -e
grep -qw bxray_tproxy /etc/passwd || echo "bxray_tproxy:x:0:23333:::" >> /etc/passwd`)
	return err
}

func GenerateRoutingTable() error {
	script := `
set -e
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

	script := []string{"set -e", "iptables -t mangle -N BXRAY"}
	for _, subnet := range subnets {
		script = append(script, fmt.Sprintf(`iptables -t mangle -A BXRAY -d %s -j RETURN`, subnet))
	}
	script = append(script, `iptables -t mangle -A BXRAY -d 224.0.0.0/3 -j RETURN`)

	_, err = runScriptWithSh(strings.Join(script, "\n"))
	return err
}

func MarkAndForwardPackets() error {
	script := `
set -e
iptables -t mangle -A BXRAY -p tcp -j TPROXY --on-port 13345 --tproxy-mark 127
iptables -t mangle -A BXRAY -p udp -j TPROXY --on-port 13345 --tproxy-mark 127
	`
	_, err := runScriptWithSh(script)
	return err
}

func ApplyMainRules() error {
	script := `
set -e
iptables -t mangle -A PREROUTING -j BXRAY`
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
		"set -e",
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
	script := `
set -e
iptables -t mangle -I BXRAY_MASK 1 -p udp --dport 53 -j RETURN
iptables -t mangle -I BXRAY_MASK 1 -p tcp --dport 53 -j RETURN
iptables -t mangle -I BXRAY 1 -p udp --dport 53 -j RETURN
iptables -t mangle -I BXRAY 1 -p tcp --dport 53 -j RETURN`
	_, err := runScriptWithSh(script)
	return err
}

func HijackDns() error {
	script := `
set -e
iptables -t mangle -I BXRAY_MASK 2 -p udp --dport 53 -j MARK --set-mark 127
iptables -t mangle -I BXRAY_MASK 2 -p tcp --dport 53 -j MARK --set-mark 127
iptables -t mangle -I BXRAY 1 -p udp --dport 53 -j TPROXY --on-port 13345 --tproxy-mark 127
iptables -t mangle -I BXRAY 1 -p tcp --dport 53 -j TPROXY --on-port 13345 --tproxy-mark 127`
	_, err := runScriptWithSh(script)
	return err
}

func LoosenRpFilters() error {
	script := `
set -e
sysctl -w net.ipv4.conf.all.rp_filter=2
sysctl -w net.ipv4.conf.default.rp_filter=2
	`
	_, err := runScriptWithSh(script)
	return err
}

func CleanUp() error {
	script := `
set -e

iptables -t mangle -S >/dev/null
ip rule show >/dev/null

delete_all_matches() {
    table=$1
    shift
    while iptables -t "$table" -C "$@" >/dev/null 2>&1; do
        iptables -t "$table" -D "$@"
    done
}

delete_all_matches mangle PREROUTING -j BXRAY
delete_all_matches mangle OUTPUT -p tcp -j BXRAY_MASK
delete_all_matches mangle OUTPUT -p udp -j BXRAY_MASK

if iptables -t mangle -S BXRAY >/dev/null 2>&1; then
    iptables -t mangle -F BXRAY
    iptables -t mangle -X BXRAY
fi
if iptables -t mangle -S BXRAY_MASK >/dev/null 2>&1; then
    iptables -t mangle -F BXRAY_MASK
    iptables -t mangle -X BXRAY_MASK
fi

while ip route del local 0.0.0.0/0 dev lo table 102 >/dev/null 2>&1; do
    :
done
while ip rule del fwmark 127 table 102 >/dev/null 2>&1; do
    :
done
`
	_, err := runScriptWithSh(script)
	return err
}
