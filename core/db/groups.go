package db

import (
	"bushuray-core/structs"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func (db *DB) DeleteGroup(id int) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	group_config_dir := db.GetGroupDirPath(id)
	err := os.RemoveAll(group_config_dir)
	if err != nil {
		return fmt.Errorf("Failed to delete group dir %w", err)
	}
	return nil
}

func (db *DB) AddGroup(name string, subscription_url string) (structs.GroupAdded, error) {
	var group_added structs.GroupAdded
	db.mu.Lock()
	defer db.mu.Unlock()
	db_config, err := db.loadDBConfig()
	if err != nil {
		return group_added, err
	}
	db_config.LastGroupId++
	err = db.saveDBConfig(db_config)
	if err != nil {
		return group_added, err
	}

	group_id := db_config.LastGroupId
	group_dir_path := db.GetGroupDirPath(group_id)
	group_config_path := db.GetGroupConfigFilePath(group_id)
	err = os.RemoveAll(group_dir_path)
	if err != nil {
		return group_added, err
	}

	err = os.MkdirAll(group_dir_path, 0755)
	if err != nil {
		return group_added, err
	}

	group := structs.Group{
		Id:              group_id,
		SubscriptionUrl: subscription_url,
		Name:            name,
		LastId:          0,
	}

	group_json, err := json.MarshalIndent(group, "", " ")
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(group_config_path, group_json, 0644)
	if err != nil {
		return group_added, fmt.Errorf("failed to write %s: %w", group_config_path, err)
	}

	group_added = structs.GroupAdded{
		Id:              group_id,
		Name:            name,
		SubscriptionUrl: subscription_url,
	}

	return group_added, nil
}


func (db *DB) loadGroupConfig(id int) (structs.Group, error) {
	var group_data structs.Group
	group_conf_file := db.GetGroupConfigFilePath(id)
	data, err := os.ReadFile(group_conf_file)
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

func (db *DB) saveGroupConfig(group structs.Group) error {
	group_conf_file := db.GetGroupConfigFilePath(group.Id)
	json_data, err := json.MarshalIndent(group, "", " ")

	if err != nil {
		log.Fatal("failed to stringify default group config")
	}

	if err := os.WriteFile(group_conf_file, json_data, 0644); err != nil {
		log.Fatal("failed to write to group config " + group_conf_file + ": " + err.Error())
	}
	return nil
}
