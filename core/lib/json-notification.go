package lib

import (
	"bushuray-core/structs"
	"encoding/json"
	"log"
)

func CreateJsonNotification(msg string, obj any) []byte {
	data := structs.Message[any]{
		Msg:  msg,
		Data: obj,
	}
	json_data, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("failed to parse json trying to send a message %v %s", data, err)
	}
	return json_data
}
