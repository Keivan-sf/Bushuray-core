package tunscripts

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func extractSubnets(rawOutput string) []string {
	subnets := strings.Split(rawOutput, "\n")
	results := []string{}

	for _, subnet := range subnets {
		trimmed := strings.TrimSpace(subnet)
		if len(trimmed) < 1 {
			continue
		}
		results = append(results, trimmed)
	}

	return results
}

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
