package mainproxy

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	tunscripts "bushuray-core/lib/proxy/mainproxy/scripts"
)

const tunStatePath = "/run/bushuray-core-tun-state.json"
const tunStateVersion = 1

type persistedTunState struct {
	Version         int    `json:"version"`
	RpFilterAll     string `json:"rpFilterAll"`
	RpFilterDefault string `json:"rpFilterDefault"`
}

func (p *ProxyManager) beginTunMode() error {
	if err := p.rollbackTunMode(); err != nil {
		return fmt.Errorf("clean previous TUN state: %w", err)
	}

	rpFilterAll, err := readRpFilter("all")
	if err != nil {
		return fmt.Errorf("read all rp_filter: %w", err)
	}
	rpFilterDefault, err := readRpFilter("default")
	if err != nil {
		return fmt.Errorf("read default rp_filter: %w", err)
	}

	if err := persistTunState(persistedTunState{
		Version:         tunStateVersion,
		RpFilterAll:     rpFilterAll,
		RpFilterDefault: rpFilterDefault,
	}); err != nil {
		return fmt.Errorf("persist TUN state: %w", err)
	}
	return nil
}

func (p *ProxyManager) recoverPersistedTunMode() error {
	if !tunStateExists() {
		return nil
	}
	return p.rollbackTunMode()
}

func (p *ProxyManager) rollbackTunMode() error {
	cleanupErr := tunscripts.CleanUp()

	state, stateErr := loadTunState()
	if errors.Is(stateErr, os.ErrNotExist) {
		return cleanupErr
	}
	if stateErr != nil {
		return errors.Join(cleanupErr, fmt.Errorf("load TUN state: %w", stateErr))
	}

	restoreErr := errors.Join(
		setRpFilter("all", state.RpFilterAll),
		setRpFilter("default", state.RpFilterDefault),
	)
	rollbackErr := errors.Join(cleanupErr, restoreErr)
	if rollbackErr != nil {
		return rollbackErr
	}
	return removeTunState()
}

func readRpFilter(scope string) (string, error) {
	output, err := exec.Command("sysctl", "-n", "net.ipv4.conf."+scope+".rp_filter").Output()
	if err != nil {
		return "", err
	}
	value := strings.TrimSpace(string(output))
	if err := validateRpFilter(value); err != nil {
		return "", err
	}
	return value, nil
}

func setRpFilter(scope string, value string) error {
	if err := validateRpFilter(value); err != nil {
		return err
	}
	return exec.Command(
		"sysctl",
		"-w",
		"net.ipv4.conf."+scope+".rp_filter="+value,
	).Run()
}

func validateRpFilter(value string) error {
	switch value {
	case "0", "1", "2":
		return nil
	default:
		return fmt.Errorf("invalid rp_filter value %q", value)
	}
}

func persistTunState(state persistedTunState) error {
	stateJSON, err := json.Marshal(state)
	if err != nil {
		return err
	}

	stateFile, err := os.CreateTemp("/run", ".bushuray-core-tun-state-*")
	if err != nil {
		return err
	}
	tempPath := stateFile.Name()
	defer os.Remove(tempPath)

	if err := stateFile.Chmod(0o600); err != nil {
		stateFile.Close()
		return err
	}
	if _, err := stateFile.Write(stateJSON); err != nil {
		stateFile.Close()
		return err
	}
	if err := stateFile.Sync(); err != nil {
		stateFile.Close()
		return err
	}
	if err := stateFile.Close(); err != nil {
		return err
	}
	return os.Rename(tempPath, tunStatePath)
}

func loadTunState() (persistedTunState, error) {
	stateJSON, err := os.ReadFile(tunStatePath)
	if err != nil {
		return persistedTunState{}, err
	}
	return decodeTunState(stateJSON)
}

func decodeTunState(stateJSON []byte) (persistedTunState, error) {
	var state persistedTunState
	if err := json.Unmarshal(stateJSON, &state); err != nil {
		return state, err
	}
	if state.Version != tunStateVersion {
		return state, fmt.Errorf("unsupported TUN state version %d", state.Version)
	}
	if err := validateRpFilter(state.RpFilterAll); err != nil {
		return state, fmt.Errorf("rpFilterAll: %w", err)
	}
	if err := validateRpFilter(state.RpFilterDefault); err != nil {
		return state, fmt.Errorf("rpFilterDefault: %w", err)
	}
	return state, nil
}

func tunStateExists() bool {
	_, err := os.Stat(tunStatePath)
	return !errors.Is(err, os.ErrNotExist)
}

func removeTunState() error {
	err := os.Remove(tunStatePath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
