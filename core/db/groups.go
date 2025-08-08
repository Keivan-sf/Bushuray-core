package db

import (
	"bushuray-core/structs"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"syscall"
)

func (db *DB) UpdateGroupAndProfiles(group_id int, profiles []structs.DBAddProfileData) ([]structs.Profile, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	profiles_added := []structs.Profile{}
	group_dir := db.GetGroupDirPath(group_id)

	group_config, err := db.loadGroupConfig(group_id)
	group_config.LastId = 0
	if err != nil {
		return profiles_added, fmt.Errorf("Error getting group: %w", err)
	}

	entries, err := os.ReadDir(group_dir)
	if err != nil {
		return profiles_added, fmt.Errorf("Error reading directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") || entry.Name() == "group_config.json" {
			continue
		}
		profile_id_str, _ := strings.CutSuffix(entry.Name(), ".json")
		profile_id, err := strconv.Atoi(profile_id_str)
		if err != nil {
			continue
		}
		db.deleteProfile(group_id, profile_id)
	}

	err = db.updateGroup(group_config)
	if err != nil {
		log.Println("warning while updating the group:", err)
	}

	for _, profile := range profiles {
		profile_added, err := db.addProfile(profile)
		if err == nil {
			profiles_added = append(profiles_added, profile_added)
		}
	}

	return profiles_added, nil
}

func (db *DB) updateGroup(group structs.Group) error {
	oldUmask := syscall.Umask(0)
	defer syscall.Umask(oldUmask)
	group_config_path := db.GetGroupConfigFilePath(group.Id)
	group_json, err := json.MarshalIndent(group, "", " ")
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(group_config_path, group_json, 0666)
	if err != nil {
		return fmt.Errorf("failed to write %s: %w", group_config_path, err)
	}
	return nil
}

func (db *DB) GetAllGroupsAndProfiles() ([]structs.GroupWithProfiles, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	var groups_with_profiles []structs.GroupWithProfiles = []structs.GroupWithProfiles{}

	dir_path := db.GetGroupsDirPath()

	entries, err := os.ReadDir(dir_path)
	if err != nil {
		return groups_with_profiles, fmt.Errorf("Error reading directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		group_id, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}
		group, err := db.getGroupDataAndProfiles(group_id)
		if err != nil {
			log.Println("warning while gathering all groups:", err)
			continue
		}
		groups_with_profiles = append(groups_with_profiles, group)
	}

	return groups_with_profiles, nil
}

func (db *DB) getGroupDataAndProfiles(group_id int) (structs.GroupWithProfiles, error) {
	var group_with_profiles structs.GroupWithProfiles = structs.GroupWithProfiles{
		Profiles: []structs.Profile{},
	}
	dir_path := db.GetGroupDirPath(group_id)
	group_config, err := db.loadGroupConfig(group_id)
	if err != nil {
		return group_with_profiles, err
	}
	group_with_profiles.Group = group_config

	entries, err := os.ReadDir(dir_path)
	if err != nil {
		return group_with_profiles, fmt.Errorf("Error reading directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		profile_id_str, _ := strings.CutSuffix(entry.Name(), ".json")
		profile_id, err := strconv.Atoi(profile_id_str)
		if err != nil {
			continue
		}
		profile, err := db.getProfile(group_id, profile_id)
		if err != nil {
			continue
		}
		group_with_profiles.Profiles = append(group_with_profiles.Profiles, profile)
	}

	return group_with_profiles, nil
}

func (db *DB) DeleteGroup(id int) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.deleteGroup(id)
}

func (db *DB) deleteGroup(id int) error {
	group_config_dir := db.GetGroupDirPath(id)
	err := os.RemoveAll(group_config_dir)
	if err != nil {
		return fmt.Errorf("Failed to delete group dir %w", err)
	}
	return nil
}

func (db *DB) AddGroup(name string, subscription_url string) (structs.GroupAdded, error) {
	oldUmask := syscall.Umask(0)
	defer syscall.Umask(oldUmask)

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

	err = os.MkdirAll(group_dir_path, 0777)
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

	err = os.WriteFile(group_config_path, group_json, 0666)
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

func (db *DB) LoadGroupConfig(id int) (structs.Group, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	return db.loadGroupConfig(id)
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
	oldUmask := syscall.Umask(0)
	defer syscall.Umask(oldUmask)

	group_conf_file := db.GetGroupConfigFilePath(group.Id)
	json_data, err := json.MarshalIndent(group, "", " ")

	if err != nil {
		log.Fatal("failed to stringify default group config")
	}

	if err := os.WriteFile(group_conf_file, json_data, 0666); err != nil {
		log.Fatal("failed to write to group config " + group_conf_file + ": " + err.Error())
	}
	return nil
}
