package mainproxy

import (
	"bushuray-core/lib"
	"bushuray-core/lib/proxy/xray"
	"bushuray-core/structs"
	"fmt"
	"net/http"
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
	sem := make(chan struct{}, 5)

	for profile := range tests_chan {
		sem <- struct{}{}
		go func(profile structs.Profile) {
			ping := p.test(profile)
			p.sendTestResult(profile, ping)
			<-sem
		}(profile)
	}
}

func (p *ProxyManager) test(profile structs.Profile) int {
	port, err := p.portPool.GetPort()
	if err != nil {
		return -1
	}
	parsed, err := lib.ParseUri(profile.Uri, port, -1)
	if err != nil {
		return -1
	}

	xray_core := xray.XrayCore{
		Exited: make(chan error),
	}

	xray_core.Start(parsed)
	defer xray_core.Stop()
	time.Sleep(1 * time.Second)

	// test a request with socks5
	dialer, err := goproxy.SOCKS5("tcp", fmt.Sprintf("127.0.0.1:%d", port), nil, goproxy.Direct)
	if err != nil {
		return -1
	}

	transport := &http.Transport{
		Dial: dialer.Dial,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   5 * time.Second,
	}
	start_time := time.Now()
	_, err = client.Get("https://cp.cloudflare.com")
	ping := time.Since(start_time)

	if err != nil {
		return -1
	}

	return int(ping.Milliseconds())
}

func (p *ProxyManager) sendTestResult(profile structs.Profile, ping int) {
	profile.TestResult = ping
	p.TestResultChannel <- TestResult{
		Profile: profile,
	}
}
