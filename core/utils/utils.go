package utils

import "os"

func IsRoot() bool {
	uid := os.Getuid()
	return uid == 0
}
