package structs

import "encoding/json"

// database types
type DBAddProfileData struct {
	Uri      string `json:"uri"`
	GroupId  int    `json:"group_id"`
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
}

// commands and responses
type DeleteProfilesData struct {
	Profiles []ProfileID `json:"profiles"`
}

type ProfilesDeleted struct {
	DeletedProfiles []ProfileID `json:"deleted-profiles"`
}

type AddProfilesData struct {
	Uris    string `json:"uris"`
	GroupId int    `json:"group_id"`
}

type ProfilesAdded struct {
	Profiles []ProfileAdded `json:"profiles"`
}

type ProfileAdded struct {
	Id       int    `json:"id"`
	GroupId  int    `json:"group_id"`
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
	Uri      string `json:"uri"`
}

type AddGroupData struct {
	Name            string `json:"name"`
	SubscriptionUrl string `json:"subscription_url"`
}

type GroupAdded struct {
	Id              int    `json:"id"`
	SubscriptionUrl string `json:"subscription_url"`
	Name            string `json:"name"`
}

type DeleteGroupData struct {
	Id int `json:"id"`
}

type GroupDeleted struct {
	Id int `json:"id"`
}

type ConnectData struct {
	Profile ProfileID `json:"profile"`
}

// general types
type DBConfig struct {
	LastGroupId int `json:"last_group_id"`
}

type ApplicationData struct {
	Groups                    []GroupWithProfiles `json:"groups"`
	CurrentlyConnectedProfile ProfileID
}

type GroupWithProfiles struct {
	Group    Group     `json:"group"`
	Profiles []Profile `json:"profiles"`
}

type Group struct {
	Id              int    `json:"id"`
	SubscriptionUrl string `json:"subscription_url"`
	Name            string `json:"name"`
	LastId          int    `json:"last_id"`
}

type Profile struct {
	Id         int    `json:"id"`
	Name       string `json:"name"`
	Protocol   string `json:"protocol"`
	Uri        string `json:"uri"`
	TestResult int    `json:"test-result"`
}

type ProfileID struct {
	Id      int `json:"id"`
	GroupId int `json:"group_id"`
}

type TCPMessage struct {
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

type Message[T any] struct {
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}

type ProxyStatus struct {
	Connection string  `json:"connection"`
	Profile    Profile `json:"profile"`
}
