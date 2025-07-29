package db

import (
	"bushuray-core/structs"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

type DB struct {
	Path string
	mu   sync.Mutex
}

func (db *DB) AddProfile(data structs.AddProfileData) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	group_data, err := db.getGroupConfig(data.GroupId)
	if err != nil {
		return err
	}
	group_data.LastId++
	profile_id := group_data.LastId
	err = os.Remove(db.GetConfigFilePath(group_data.Id, profile_id))
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("error is ok: not found")
		}
		return err
	}
	fmt.Println("removed the file")


	return nil
}

func (db *DB) getGroupConfig(id int) (structs.Group, error) {
	var group_data structs.Group
	group_conf_dir := db.GetGroupConfigFilePath(id)
	data, err := os.ReadFile(group_conf_dir)
	if err != nil {
		return group_data, err
	}
	err = json.Unmarshal(data, &group_data)
	if err != nil {
		err = fmt.Errorf("Failed to parse json group data for %d: %w", id, err)
		return group_data, err
	}
	return group_data, nil
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

	group := structs.Group{
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

func (db *DB) GetGroupConfigFilePath(group_id int) string {
	return filepath.Join(db.Path, "groups", strconv.Itoa(group_id), "group_config.json")
}

func (db *DB) GetConfigFilePath(group_id int, profile_id int) string {
	return filepath.Join(db.Path, "groups", strconv.Itoa(group_id), fmt.Sprintf("%d.json", profile_id))
}
