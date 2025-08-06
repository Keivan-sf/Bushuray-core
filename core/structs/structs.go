package structs

import "encoding/json"

// database types
type DBAddProfileData struct {
	Uri      string `json:"uri"`
	GroupId  int    `json:"group_id"`
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
	Address  string `json:"address,omitzero"`
	Host     string `json:"host,omitzero"`
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
	Profiles []Profile `json:"profiles"`
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

type DisconnectData struct{}

type TestProfileData struct {
	Profile ProfileID `json:"profile"`
}

type ProfileUpdated struct {
	Profile Profile `json:"profile"`
}

type GetApplicationStateData struct{}

type ApplicationState struct {
	Groups           []GroupWithProfiles `json:"groups"`
	ConnectionStatus ProxyStatus         `json:"connection-status"`
}

type UpdateSubscriptionData struct {
	GroupId int `json:"group_id"`
}

type SubscriptionUpdated struct {
	GroupId  int       `json:"group_id"`
	Profiles []Profile `json:"profiles"`
}

type Warning struct {
	Key     string `json:"key"`
	Content string `json:"content"`
}

type DieData struct{}

type IsRootData struct{}
type IsRootAnswer struct {
	IsRoot bool
}

// general types
type DBConfig struct {
	LastGroupId int `json:"last_group_id"`
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
	GroupId    int    `json:"group_id"`
	Name       string `json:"name"`
	Protocol   string `json:"protocol"`
	Uri        string `json:"uri"`
	Address    string `json:"address,omitzero"`
	Host       string `json:"host,omitzero"`
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

type DisableTunData struct{}
type EnableTunData struct{}

type TunStatus struct {
	IsEnabled bool `json:"is_enabled"`
}
