package db

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

type Group struct {
	Id              int    `json:"id"`
	SubscriptionUrl string `json:"subscription_url"`
	Name            string `json:"name"`
	LastId          int    `json:"last_id"`
}

type DB struct {
	Path string
	mu   sync.Mutex
}

func (db *DB) AddConfig() {
	groupData, err := db.getGroupConfig(0)
	if err != nil {
		panic(err)
	}
	fmt.Println(groupData)
}

func (db *DB) getGroupConfig(id int) (Group, error) {
	fmt.Println("db path is in get group config is:", db.Path)
	var groupData Group
	groupConfDir := filepath.Join(db.Path, "groups", strconv.Itoa(id), "group_config.json")
	data, err := os.ReadFile(groupConfDir)
	if err != nil {
		return groupData, err
	}
	err = json.Unmarshal(data, &groupData)
	if err != nil {
		err = fmt.Errorf("Failed to parse json group data for %d: %w", id, err)
		return groupData, err
	}
	return groupData, nil
}

func (db *DB) Initialize() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic("cannot get user home directory")
	}
	var dbPath = filepath.Join(homeDir, ".config", "bushuray", "db")
	db.Path = dbPath
	var dirPath = filepath.Join(homeDir, ".config", "bushuray", "db", "groups", "0")
	filePath := filepath.Join(dirPath, "group_config.json")

	if _, err := os.Stat(filePath); err == nil {
		fmt.Println("using existing database")
		return
	} else if !os.IsNotExist(err) {
		panic("error checking for database path " + filePath + ": " + err.Error())
	}

	if err := os.MkdirAll(dirPath, 0755); err != nil {
		panic("failed to create database directory " + dirPath + ": " + err.Error())
	}

	group := Group{
		Id:              0,
		Name:            "Default",
		SubscriptionUrl: "",
		LastId:          0,
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
