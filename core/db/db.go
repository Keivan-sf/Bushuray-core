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

func (db *DB) DeleteProfile(group_id int, id int) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	profile_config_path := db.GetProfileFilePath(group_id, id)
	err := os.Remove(profile_config_path)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Profile removal error: %s did not exist in the first place", profile_config_path)
		}
		return fmt.Errorf("Failed to delete config file %w", err)
	}
	return nil
}

func (db *DB) AddProfile(data structs.DBAddProfileData) (structs.ProfileAdded, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	var profile_added structs.ProfileAdded
	group_data, err := db.loadGroupConfig(data.GroupId)
	if err != nil {
		return profile_added, err
	}
	group_data.LastId++
	err = db.saveGroupConfig(group_data)
	if err != nil {
		return profile_added, err
	}

	profile_id := group_data.LastId
	profile_path := db.GetProfileFilePath(group_data.Id, profile_id)
	err = os.Remove(db.GetProfileFilePath(group_data.Id, profile_id))
	if err != nil && !os.IsNotExist(err) {
		return profile_added, err
	}
	profile := structs.Profile{
		Id:       profile_id,
		Name:     data.Name,
		Protocol: data.Protocol,
		Uri:      data.Uri,
	}
	profile_json, err := json.MarshalIndent(profile, "", " ")
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(profile_path, profile_json, 0644)
	if err != nil {
		return profile_added, fmt.Errorf("failed to write %s: %w", profile_path, err)
	}

	profile_added = structs.ProfileAdded{
		Uri:      data.Uri,
		GroupId:  data.GroupId,
		Id:       profile_id,
		Protocol: profile.Protocol,
		Name:     profile.Name,
	}

	return profile_added, nil
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

func (db *DB) saveDBConfig(db_config structs.DBConfig) error {
	db_config_file := db.GetDBConfigFile()
	json_data, err := json.MarshalIndent(db_config, "", " ")

	if err != nil {
		log.Fatal("failed to stringify db config")
	}

	if err := os.WriteFile(db_config_file, json_data, 0644); err != nil {
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

func (db *DB) Initialize() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("cannot get user home directory")
	}
	var db_path = filepath.Join(homeDir, ".config", "bushuray", "db")
	db.Path = db_path
	if err := os.MkdirAll(db_path, 0755); err != nil {
		log.Fatal("failed to create database directory " + db_path + ": " + err.Error())
	}
	db.ensureDBConfigExistance()
	db.ensureDefaultGroupExistance()
}

func (db *DB) ensureDefaultGroupExistance() {
	var default_group_dir = filepath.Join(db.Path, "groups", "0")

	if err := os.MkdirAll(default_group_dir, 0755); err != nil {
		log.Fatal("failed to create default group directory " + default_group_dir + ": " + err.Error())
	}

	default_group_path := db.GetGroupConfigFilePath(0)
	if _, err := os.Stat(default_group_path); err == nil {
		fmt.Println("using existing default group")
		return
	} else if !os.IsNotExist(err) {
		log.Fatal("error checking for default groupo path " + default_group_path + ": " + err.Error())
	}

	group := structs.Group{
		Id:              0,
		Name:            "Default",
		SubscriptionUrl: "",
		LastId:          0,
	}

	json_data, err := json.MarshalIndent(group, "", " ")

	if err != nil {
		log.Fatal("failed to stringify default group config")
	}

	if err := os.WriteFile(default_group_path, json_data, 0644); err != nil {
		log.Fatal("failed to write to default config " + default_group_path + ": " + err.Error())
	}

	fmt.Println("default group config initialized:", default_group_path)
}

func (db *DB) ensureDBConfigExistance() {
	db_config_path := db.GetDBConfigFile()

	if _, err := os.Stat(db_config_path); err == nil {
		fmt.Println("using existing config")
	} else if !os.IsNotExist(err) {
		log.Fatal("error checking for database config " + db_config_path + ": " + err.Error())
	} else {
		db_config := structs.DBConfig{}
		json_data, err := json.MarshalIndent(db_config, "", " ")

		if err != nil {
			log.Fatal("failed to stringify db config config")
		}

		if err := os.WriteFile(db_config_path, json_data, 0644); err != nil {
			log.Fatal("failed to write to db config " + db_config_path + ": " + err.Error())
		}
	}

}

func (db *DB) GetDBConfigFile() string {
	return filepath.Join(db.Path, "config.json")
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
