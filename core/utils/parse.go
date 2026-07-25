package utils

import (
	"fmt"
	"os/exec"
	"strconv"
)

func ParseUri(uri string, socksport int, httpport int, tproxyport int) ([]byte, error) {
	var parsed_config []byte
	v2parserbin, err := GetV2parserBin()
	if err != nil {
		return parsed_config, fmt.Errorf("failed to parse: %w", err)
	}
	var v2parser_parse_cmd *exec.Cmd

	args := []string{uri, "--socksport", strconv.Itoa(socksport)}
	if httpport != -1 {
		args = append(args, "--httpport", strconv.Itoa(httpport))
	}
	if tproxyport != -1 {
		args = append(args, "--tproxyport", strconv.Itoa(tproxyport))
	}

	v2parser_parse_cmd = exec.Command(v2parserbin, args...)

	parsed_config, err = v2parser_parse_cmd.Output()
	if err != nil {
		return parsed_config, fmt.Errorf("parsing uri fialed: %w", err)
	}

	return parsed_config, nil
}
