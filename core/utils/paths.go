package utils

import (
	"fmt"
	"log"
	"os"
	osuser "os/user"
	"path/filepath"
	"strconv"
)

func GetV2parserBin() (string, error) {
	return getBin("v2parser")
}

func GetTun2socksBin() (string, error) {
	return getBin("tun2socks")
}

func GetXrayBin() (string, error) {
	return getBin("xray")
}

func getBin(name string) (string, error) {
	bin_path := filepath.Join(GetWorkingDir(), "bin", name)
	if fileExists(bin_path) {
		return bin_path, nil
	}

	homeDir, err := GetHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	localBinPath := filepath.Join(homeDir, ".local", "share", "bushuray", "bin", name)
	if fileExists(localBinPath) {
		return localBinPath, nil
	}
	return "", fmt.Errorf("could not find %v in %v nor in %v", name, bin_path, localBinPath)
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil && !info.IsDir()
}

func GetWorkingDir() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return dir
}

func GetHomeDir() (string, error) {
	uid := os.Getuid()
	if uid == 0 {
		real_uid, err := strconv.Atoi(os.Getenv("SUDO_UID"))
		if err != nil {
			return "", fmt.Errorf("failed to get user id outside of sudo %w", err)
		}
		uid = real_uid
	}

	user, err := osuser.LookupId(strconv.Itoa(uid))
	if err != nil {
		log.Fatal("failed to get user from uid")
		return "", fmt.Errorf("failed to get user from uid %d: %w", uid, err)
	}
	log.Printf("using home directory:%v\n", user.HomeDir)
	return user.HomeDir, nil
}
