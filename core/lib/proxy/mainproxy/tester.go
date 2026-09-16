package mainproxy

import (
	"fmt"
	"net/http"
	"time"

	"bushuray-core/lib/proxy/xray"
	"bushuray-core/structs"
	"bushuray-core/utils"

	goproxy "golang.org/x/net/proxy"
)

type TestResult struct {
	Success    bool
	Profile    structs.Profile
	Generation uint64
}

type TestRequest struct {
	Profile    structs.Profile
	Generation uint64
}

func (p *ProxyManager) TestProfile(profile structs.Profile) {
	p.testChannel <- TestRequest{
		Profile:    profile,
		Generation: p.testGeneration.Load(),
	}
}

func (p *ProxyManager) StopTests() {
	p.testGeneration.Add(1)
}

func (p *ProxyManager) IsCurrentTestGeneration(generation uint64) bool {
	return p.testGeneration.Load() == generation
}

func (p *ProxyManager) listenForTests(tests_chan chan TestRequest) {
	sem := make(chan struct{}, 5)

	for req := range tests_chan {
		if !p.IsCurrentTestGeneration(req.Generation) {
			continue
		}
		sem <- struct{}{}
		if !p.IsCurrentTestGeneration(req.Generation) {
			<-sem
			continue
		}
		go func(req TestRequest) {
			defer func() { <-sem }()
			ping := p.test(req.Profile)
			p.sendTestResult(req.Profile, ping, req.Generation)
		}(req)
	}
}

func (p *ProxyManager) test(profile structs.Profile) int {
	port, err := p.portPool.GetPort()
	if err != nil {
		return -1
	}
	defer p.portPool.ReleasePort(port)
	parsed, err := utils.ParseUri(profile.Uri, port, -1, -1)
	if err != nil {
		return -1
	}

	xray_core := xray.XrayCore{
		Exited: make(chan error, 1),
	}

	err = xray_core.Start(parsed)
	if err != nil {
		return -1
	}

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
	_, err = client.Get(p.appConfig.TestURL)
	ping := time.Since(start_time)

	if err != nil {
		return -1
	}

	return int(ping.Milliseconds())
}

func (p *ProxyManager) sendTestResult(profile structs.Profile, ping int, generation uint64) {
	if !p.IsCurrentTestGeneration(generation) {
		return
	}
	profile.TestResult = ping
	p.TestResultChannel <- TestResult{
		Profile:    profile,
		Generation: generation,
	}
}
