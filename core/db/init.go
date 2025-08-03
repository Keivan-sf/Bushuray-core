package db

import (
	"bushuray-core/utils"
	"bushuray-core/structs"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

func (db *DB) Initialize() {
	homeDir, err := utils.GetHomeDir()
	if err != nil {
		log.Fatal("cannot get user home directory")
	}
	var db_path = filepath.Join(homeDir, ".config", "bushuray", "db")
	db.Path = db_path
	if err := os.MkdirAll(db_path, 0777); err != nil {
		log.Fatal("failed to create database directory " + db_path + ": " + err.Error())
	}
	db.ensureDBConfigExistance()
	db.ensureDefaultGroupExistance()
}

func (db *DB) ensureDefaultGroupExistance() {
	var default_group_dir = filepath.Join(db.Path, "groups", "0")

	if err := os.MkdirAll(default_group_dir, 0777); err != nil {
		log.Fatal("failed to create default group directory " + default_group_dir + ": " + err.Error())
	}

	default_group_path := db.GetGroupConfigFilePath(0)
	if _, err := os.Stat(default_group_path); err == nil {
		log.Println("using existing default group")
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

	if err := os.WriteFile(default_group_path, json_data, 0666); err != nil {
		log.Fatal("failed to write to default config " + default_group_path + ": " + err.Error())
	}

	log.Println("default group config initialized:", default_group_path)
}

func (db *DB) ensureDBConfigExistance() {
	db_config_path := db.GetDBConfigFile()

	if _, err := os.Stat(db_config_path); err == nil {
		log.Println("using existing config")
	} else if !os.IsNotExist(err) {
		log.Fatal("error checking for database config " + db_config_path + ": " + err.Error())
	} else {
		db_config := structs.DBConfig{}
		json_data, err := json.MarshalIndent(db_config, "", " ")

		if err != nil {
			log.Fatal("failed to stringify db config config")
		}

		if err := os.WriteFile(db_config_path, json_data, 0666); err != nil {
			log.Fatal("failed to write to db config " + db_config_path + ": " + err.Error())
		}
	}

}
