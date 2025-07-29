package structs

// database types
type DBAddProfileData struct {
	Uri        string          `json:"uri"`
	GroupId    int             `json:"group_id"`
	Name       string          `json:"name"`
	Protocol   string          `json:"protocol"`
}

// commands and responses
type AddProfileData struct {
	Uri     string `json:"uri"`
	GroupId int    `json:"group_id"`
}

type ProfileAdded struct {
	Id       int    `json:"id"`
	GroupId  int    `json:"group_id"`
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
	Uri      string `json:"uri"`
}

// general types
type Group struct {
	Id              int    `json:"id"`
	SubscriptionUrl string `json:"subscription_url"`
	Name            string `json:"name"`
	LastId          int    `json:"last_id"`
}

type Profile struct {
	Id         int             `json:"id"`
	Name       string          `json:"name"`
	Protocol   string          `json:"protocol"`
	Uri        string          `json:"uri"`
}
