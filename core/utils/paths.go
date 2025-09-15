package utils

import (
	"fmt"
	"log"
	"os"
	osuser "os/user"
	"strconv"
)

func GetV2parserBin() (string, error) {
	return GetBinPath("v2parser")
}

func GetTun2socksBin() (string, error) {
	return GetBinPath("tun2socks")
}

func GetXrayBin() (string, error) {
	return GetBinPath("xray")
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
	return user.HomeDir, nil
}

func RemoveDuplicates(input []string) []string {
	seen := make(map[string]struct{})
	result := []string{}

	for _, v := range input {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}
