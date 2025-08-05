package utils

import (
	"context"
	"fmt"
	"net"
	"time"
)

func ResolveDomainIpv4(domain string) (string, error) {
	resolved, err := ResolveDomain(domain)
	if err != nil {
		return "", err
	}

	ipv4 := ""
	for _, ip := range resolved {
		if ip.To4() != nil {
			ipv4 = ip.String()
			break
		}
	}

	if ipv4 == "" {
		return ipv4, fmt.Errorf("no ipv4 found for domain: %s", domain)
	}
	return ipv4, nil
}

func ResolveDomain(domain string) ([]net.IP, error) {
	dialer := &net.Dialer{}
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			return dialer.DialContext(ctx, "udp", "1.1.1.1:53")
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	ips, err := resolver.LookupIP(ctx, "ip", domain)
	if err != nil {
		return ips, nil
	}

	ips, err = net.LookupIP(domain)
	if err == nil && len(ips) > 0 {
		return ips, nil
	}

	return nil, fmt.Errorf("failed to resolve domain: %w", err)
}
