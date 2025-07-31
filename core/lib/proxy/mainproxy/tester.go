package mainproxy

import (
	"bushuray-core/lib"
	"bushuray-core/lib/proxy/xray"
	"bushuray-core/structs"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	goproxy "golang.org/x/net/proxy"
)

type TestResult struct {
	Success bool
	Profile structs.Profile
}

func (p *ProxyManager) TestProfile(profile structs.Profile) {
	p.testChannel <- profile
}

func (p *ProxyManager) listenForTests(tests_chan chan structs.Profile) {
	for profile := range tests_chan {
		go p.test(profile)
	}
}

func (p *ProxyManager) test(profile structs.Profile) {
	port, err := p.portPool.GetPort()
	if err != nil {
		p.sendTestResult(profile, -1)
		return
	}
	parsed, err := lib.ParseUri(profile.Uri, port, -1)
	if err != nil {
		p.sendTestResult(profile, -1)
		return
	}

	xray_core := xray.XrayCore{
		Exited: make(chan error),
	}
	go func() {
		for {
			_, ok := <-xray_core.Exited
			if !ok {
				return
			}
			p.sendTestResult(profile, -1)
		}
	}()
	xray_core.Start(parsed)
	defer xray_core.Stop()
	time.Sleep(1 * time.Second)

	// test a request with socks5
	dialer, err := goproxy.SOCKS5("tcp", fmt.Sprintf("127.0.0.1:%d", port), nil, goproxy.Direct)
	if err != nil {
		p.sendTestResult(profile, -1)
		return
	}

	transport := &http.Transport{
		Dial: dialer.Dial,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   5 * time.Second,
	}
	start_time := time.Now()
	resp, err := client.Get("https://dns.google.com/resolve?name=google.com")
	if err != nil {
		p.sendTestResult(profile, -1)
		return
	}
	ping := time.Since(start_time)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		p.sendTestResult(profile, -1)
		return
	}

	bodyStr := strings.TrimSpace(string(body))

	if bodyStr == "" {
		p.sendTestResult(profile, -1)
		return
	}
	p.sendTestResult(profile, int(ping.Milliseconds()))
}

func (p *ProxyManager) sendTestResult(profile structs.Profile, ping int) {
	profile.TestResult = ping
	p.TestResultChannel <- TestResult{
		Profile: profile,
	}
}
