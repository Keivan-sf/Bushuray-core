package tunmode

import (
	"bytes"
	"fmt"
	"log"
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

func setupDnsHijackRules(interface_name string, interface_ip string, dns string) error {
	script := fmt.Sprintf(`
IFACE="%s"
IFACE_IP="%s"
DNS_IP="%s"

iptables -t nat -A OUTPUT -o "$IFACE" -p udp --dport 53 -j DNAT --to-destination "$DNS_IP":53
iptables -t nat -A OUTPUT -o "$IFACE" -p tcp --dport 53 -j DNAT --to-destination "$DNS_IP":53

ip route add $DNS_IP via $IFACE_IP

echo "DNS hijack set: all DNS over $IFACE will go to $DNS_IP"
	`, interface_name, interface_ip, dns)
	_, err := runScriptWithSh(script)
	return err

}

func cleanDnsHijackRules(interface_name string, interface_ip string, dns string) error {
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

ip route del $DNS_IP via $IFACE_IP

echo "All DNS hijack rules removed from OUTPUT chain for interface $IFACE"
	`, interface_name, interface_ip, dns)
	_, err := runScriptWithSh(script)
	return err
}
