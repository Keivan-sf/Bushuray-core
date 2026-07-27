package tunscripts

import (
	"fmt"
	"strings"
	"time"
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
		"iptables -t mangle -A BXRAY_MASK -m owner --gid-owner 23333 -j RETURN",
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
	// Position 1 in BXRAY_MASK is reserved for the Xray process GID. DNS
	// marking must come before LAN bypasses so requests to a local resolver
	// such as 127.0.0.53 or a gateway address are intercepted as well.
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

func WaitForTProxyListener(tproxyPort int, timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)
	var lastErr error

	for time.Now().Before(deadline) {
		diagnostics, err := inspectTProxyListener(tproxyPort)
		if err == nil {
			return diagnostics, nil
		}
		lastErr = err
		time.Sleep(100 * time.Millisecond)
	}

	return "", fmt.Errorf("Xray TProxy listener did not become ready within %s: %w", timeout, lastErr)
}

func inspectTProxyListener(tproxyPort int) (string, error) {
	script := fmt.Sprintf(`
set -e

ss -H -lnt | awk '$4 ~ /:%d$/ { found=1 } END { exit !found }' || {
    echo "Xray is not listening for TCP on port %d" >&2
    exit 1
}
ss -H -lnu | awk '$4 ~ /:%d$/ { found=1 } END { exit !found }' || {
    echo "Xray is not listening for UDP on port %d" >&2
    exit 1
}

echo "tcp,udp:%d"
`, tproxyPort, tproxyPort, tproxyPort, tproxyPort, tproxyPort)
	return runScriptWithSh(script)
}

func WaitForTunReady(tproxyPort int, timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)
	var lastErr error

	for time.Now().Before(deadline) {
		diagnostics, err := inspectTun(tproxyPort)
		if err == nil {
			return diagnostics, nil
		}
		lastErr = err
		time.Sleep(100 * time.Millisecond)
	}

	return "", fmt.Errorf("TProxy did not become ready within %s: %w", timeout, lastErr)
}

func inspectTun(tproxyPort int) (string, error) {
	script := fmt.Sprintf(`
set -e

ip -4 rule show | grep -Eq 'fwmark (0x)?0*7f.*lookup 102' || {
    echo "missing policy rule: fwmark 127 lookup 102" >&2
    exit 1
}
ip -4 route show table 102 2>/dev/null | grep -Eq '^local (default|0\.0\.0\.0/0) dev lo' || {
    echo "missing local route in table 102" >&2
    exit 1
}
iptables -t mangle -C PREROUTING -j BXRAY >/dev/null 2>&1 || {
    echo "missing PREROUTING -> BXRAY hook" >&2
    exit 1
}
iptables -t mangle -C OUTPUT -p tcp -j BXRAY_MASK >/dev/null 2>&1 || {
    echo "missing TCP OUTPUT -> BXRAY_MASK hook" >&2
    exit 1
}
iptables -t mangle -C OUTPUT -p udp -j BXRAY_MASK >/dev/null 2>&1 || {
    echo "missing UDP OUTPUT -> BXRAY_MASK hook" >&2
    exit 1
}
ss -H -lnt | awk '$4 ~ /:%d$/ { found=1 } END { exit !found }' || {
    echo "Xray is not listening for TCP on port %d" >&2
    exit 1
}
ss -H -lnu | awk '$4 ~ /:%d$/ { found=1 } END { exit !found }' || {
    echo "Xray is not listening for UDP on port %d" >&2
    exit 1
}

echo "fwmark=127 table=102 hooks=BXRAY/BXRAY_MASK tproxy=tcp,udp:%d"
`, tproxyPort, tproxyPort, tproxyPort, tproxyPort, tproxyPort)
	return runScriptWithSh(script)
}

func CleanUp() error {
	script := `
set -e

# Fail instead of silently claiming success when netfilter is unavailable or
# the process has insufficient privileges.
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
