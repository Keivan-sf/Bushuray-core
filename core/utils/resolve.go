package utils

import (
	"fmt"
	"log"
	"net"
	"sort"
	"time"

	"github.com/miekg/dns"
)

func ResolveDomainIpv4(domain string) ([]string, error) {
	ipSet := make(map[string]bool)
	dnsServers := []string{"1.1.1.1:53", "8.8.8.8:53", "208.67.222.222:53"}
	for _, server := range dnsServers {
		ips, err := queryDNSServer(domain, server)
		if err != nil {
			log.Printf("Query to %s failed: %v", server, err)
			continue
		}

		for _, ip := range ips {
			ipSet[ip] = true
		}
	}

	var result []string
	for ip := range ipSet {
		result = append(result, ip)
	}

	sort.Strings(result)

	if len(result) == 0 {
		log.Printf("No IPs found via DNS servers, falling back to system resolver")
		systemIPs, err := ResolveDomainSystem(domain)
		if err != nil {
			return nil, fmt.Errorf("DNS servers and system resolver both failed: %w", err)
		}
		return systemIPs, nil
	}

	return result, nil
}

func queryDNSServer(domain, server string) ([]string, error) {
	var ips []string

	udpClient := &dns.Client{
		Net:     "udp",
		Timeout: 5 * time.Second,
	}

	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeA)
	msg.RecursionDesired = true
	msg.SetEdns0(4096, false) // Disable DNS cookies

	resp, _, err := udpClient.Exchange(msg, server)
	if err != nil {
		return nil, fmt.Errorf("UDP query failed: %w", err)
	}

	if resp.Truncated {
		tcpClient := &dns.Client{
			Net:     "tcp",
			Timeout: 10 * time.Second,
		}
		resp, _, err = tcpClient.Exchange(msg, server)
		if err != nil {
			return nil, fmt.Errorf("TCP retry failed: %w", err)
		}
	}

	for _, answer := range resp.Answer {
		if a, ok := answer.(*dns.A); ok {
			ips = append(ips, a.A.String())
		}
	}

	return ips, nil
}

func ResolveDomainSystem(domain string) ([]string, error) {
	ips, err := net.LookupIP(domain)
	if err != nil {
		return nil, err
	}

	var ipv4s []string
	for _, ip := range ips {
		if ip.To4() != nil {
			ipv4s = append(ipv4s, ip.String())
		}
	}

	if len(ipv4s) == 0 {
		return ipv4s, fmt.Errorf("no ipv4 found for domain: %s", domain)
	}

	return ipv4s, nil
}
