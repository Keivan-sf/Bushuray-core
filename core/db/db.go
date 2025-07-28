package db

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Group struct {
	Id              uint   `json:"id"`
	SubscriptionUrl string `json:"subscription_url"`
	Name            string `json:"name"`
}

func Initialize() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic("cannot get user home directory")
	}

	dirPath := filepath.Join(homeDir, ".config", "bushuray", "db", "groups", "0")
	filePath := filepath.Join(dirPath, "group_config.json")

	if _, err := os.Stat(filePath); err == nil {
		return
	} else if !os.IsNotExist(err) {
		panic("error checking for database path " + filePath + ": " + err.Error())
	}

	if err := os.MkdirAll(dirPath, 0755); err != nil {
		panic("failed to create database directory " + dirPath + ": " + err.Error())
	}

	group := Group{
		Id:   0,
		Name: "Default",
	}

	json_data, err := json.MarshalIndent(group, "", " ")

	if err != nil {
		panic("failed to stringify default group config")
	}

	if err := os.WriteFile(filePath, json_data, 0644); err != nil {
		panic("failed to write to default config " + filePath + ": " + err.Error())
	}

	fmt.Println("default group config initialized:", filePath)
}
