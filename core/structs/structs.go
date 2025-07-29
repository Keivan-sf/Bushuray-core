package structs

type AddProfileData struct {
	Uri     string `json:"uri"`
	GroupId int    `json:"group_id"`
}


type Group struct {
	Id              int    `json:"id"`
	SubscriptionUrl string `json:"subscription_url"`
	Name            string `json:"name"`
	LastId          int    `json:"last_id"`
}

