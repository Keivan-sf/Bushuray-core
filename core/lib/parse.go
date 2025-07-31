package lib

import (
	"fmt"
	"os/exec"
	"path"
	"strconv"
)

func ParseUri(uri string, socksport int, httpport int) ([]byte, error) {
	var parsed_config []byte
	v2parserbin := path.Join(GetWorkingDir(), "bin", "v2parser")
	var v2parser_parse_cmd *exec.Cmd
	if httpport != -1 {
		v2parser_parse_cmd = exec.Command(v2parserbin, uri, "--socksport", strconv.Itoa(socksport), "--httpport", strconv.Itoa(httpport))
	} else {
		v2parser_parse_cmd = exec.Command(v2parserbin, uri, "--socksport", strconv.Itoa(socksport))
	}

	parsed_config, err := v2parser_parse_cmd.Output()
	if err != nil {
		return parsed_config, fmt.Errorf("parsing uri fialed: %w", err)
	}

	return parsed_config, nil
}
