package db

import (
	"bushuray-core/structs"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

type DB struct {
	Path string
	mu   sync.Mutex
}

func (db *DB) saveDBConfig(db_config structs.DBConfig) error {
	db_config_file := db.GetDBConfigFile()
	json_data, err := json.MarshalIndent(db_config, "", " ")

	if err != nil {
		log.Fatal("failed to stringify db config")
	}

	if err := os.WriteFile(db_config_file, json_data, 0666); err != nil {
		log.Fatal("failed to write to db config " + db_config_file + ": " + err.Error())
	}
	return nil
}

func (db *DB) loadDBConfig() (structs.DBConfig, error) {
	var db_config_data structs.DBConfig
	db_config_file := db.GetDBConfigFile()
	data, err := os.ReadFile(db_config_file)
	if err != nil {
		return db_config_data, err
	}

	err = json.Unmarshal(data, &db_config_data)

	if err != nil {
		err = fmt.Errorf("Failed to parse json db config: %w", err)
		return db_config_data, err
	}

	return db_config_data, nil
}

func (db *DB) GetDBConfigFile() string {
	return filepath.Join(db.Path, "config.json")
}

func (db *DB) GetGroupsDirPath() string {
	return filepath.Join(db.Path, "groups")
}

func (db *DB) GetGroupDirPath(group_id int) string {
	return filepath.Join(db.Path, "groups", strconv.Itoa(group_id))
}

func (db *DB) GetGroupConfigFilePath(group_id int) string {
	return filepath.Join(db.Path, "groups", strconv.Itoa(group_id), "group_config.json")
}

func (db *DB) GetProfileFilePath(group_id int, profile_id int) string {
	return filepath.Join(db.Path, "groups", strconv.Itoa(group_id), fmt.Sprintf("%d.json", profile_id))
}
