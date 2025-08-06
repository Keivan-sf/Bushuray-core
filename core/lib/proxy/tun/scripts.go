package tunmode

import (
	"bytes"
	"fmt"
	"os/exec"
)

func runScriptWithSh(script string) (string, error) {
	cmd := exec.Command("sh")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("failed to get stdin of sh %w", err)
	}
	_, err = stdin.Write([]byte(script))

	if err != nil {
		return "", fmt.Errorf("failed to write to stdin of sh %w", err)
	}
	err = stdin.Close()
	if err != nil {
		return "", fmt.Errorf("failed to close stdin of sh %w", err)
	}
	stderr := bytes.Buffer{}
	cmd.Stderr = &stderr
	output, err := cmd.Output()

	if err != nil {
		return "", fmt.Errorf("%w: %s", err, string(stderr.String()))
	}
	return string(output), nil
}

func setupDnsHijackRules(interface_name string, interface_ip string, dns_ip string) error {
	script := fmt.Sprintf(`
IFACE="%s"
IFACE_IP="%s"
DNS_IP="%s"

iptables -t nat -A OUTPUT -o "$IFACE" -p udp --dport 53 -j DNAT --to-destination "$DNS_IP":53
iptables -t nat -A OUTPUT -o "$IFACE" -p tcp --dport 53 -j DNAT --to-destination "$DNS_IP":53
	`, interface_name, interface_ip, dns_ip)
	_, err := runScriptWithSh(script)
	return err

}

func cleanDnsHijackRules(interface_name string, interface_ip string, dns_ip string) error {
	script := fmt.Sprintf(`
IFACE="%s"
IFACE_IP="%s"
DNS_IP="%s"

while iptables -t nat -C OUTPUT -o "$IFACE" -p udp --dport 53 -j DNAT --to-destination "$DNS_IP":53 2>/dev/null; do
     iptables -t nat -D OUTPUT -o "$IFACE" -p udp --dport 53 -j DNAT --to-destination "$DNS_IP":53
done

while iptables -t nat -C OUTPUT -o "$IFACE" -p tcp --dport 53 -j DNAT --to-destination "$DNS_IP":53 2>/dev/null; do
     iptables -t nat -D OUTPUT -o "$IFACE" -p tcp --dport 53 -j DNAT --to-destination "$DNS_IP":53
done
	`, interface_name, interface_ip, dns_ip)
	_, err := runScriptWithSh(script)
	return err
}

func createTun(name string, ip string) error {
	script := fmt.Sprintf(`
set -e
TUN_NAME="%s"
TUN_IP="%s"
ip tuntap add mode tun dev $TUN_NAME 
ip addr add $TUN_IP dev $TUN_NAME
ip link set dev $TUN_NAME up
	`, name, ip)
	_, err := runScriptWithSh(script)
	return err
}

func deleteTun(name string) error {
	script := fmt.Sprintf(`
TUN_NAME="%s"
sudo ip tuntap del mode tun dev $TUN_NAME
	`, name)
	_, err := runScriptWithSh(script)
	return err
}

func setupDnsIpRoute(dns_ip string, interface_ip string) error {
	script := fmt.Sprintf(`
set -e
IFACE_IP="%s"
DNS_IP="%s"
ip route add $DNS_IP via $IFACE_IP
	`, interface_ip, dns_ip)
	_, err := runScriptWithSh(script)
	return err
}

func deleteDnsIpRoute(dns_ip string, interface_ip string) error {
	script := fmt.Sprintf(`
IFACE_IP="%s"
DNS_IP="%s"
ip route del $DNS_IP via $IFACE_IP || true
	`, interface_ip, dns_ip)
	_, err := runScriptWithSh(script)
	return err
}

func setupProxyIpRoutes(proxy_ipv4s []string, default_interface_ip string) error {
	script := "\nset -e\n"
	for _, ip := range proxy_ipv4s {
		script += fmt.Sprintf("ip route add %s via %s\n", ip, default_interface_ip)
	}
	_, err := runScriptWithSh(script)
	return err
}

func deleteProxyIpRoutes(proxy_ipv4s []string, default_interface_ip string) error {
	script := "\n"
	for _, ip := range proxy_ipv4s {
		script += fmt.Sprintf("ip route del %s via %s || true\n", ip, default_interface_ip)
	}
	_, err := runScriptWithSh(script)
	return err
}

func setupTunIpRoute(tun_name string, tun_interface_ip string) error {
	script := fmt.Sprintf(`
set -e
TUN_NAME="%s"
TUN_IP="%s"
ip route add default via $TUN_IP dev $TUN_NAME metric 1
	`, tun_name, tun_interface_ip)

	_, err := runScriptWithSh(script)
	return err
}

func deleteTunIpRoute(tun_name string, tun_interface_ip string) error {
	script := fmt.Sprintf(`
TUN_NAME="%s"
TUN_IP="%s"
ip route del default via $TUN_IP dev $TUN_NAME metric 1 || true
	`, tun_name, tun_interface_ip)

	_, err := runScriptWithSh(script)
	return err
}
