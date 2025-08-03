package main

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"io"
	"log"
	"net"
)

type TcpMessage struct {
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

type Message[T any] struct {
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}

type DieData struct{}

type UpdateSubscriptionData struct {
	GroupId int `json:"group_id"`
}

type GetApplicationStateData struct{}

type TestProfileData struct {
	Profile ProfileID `json:"profile"`
}

type DisconnectData struct{}

type ConnectData struct {
	Profile ProfileID `json:"profile"`
}

type DeleteGroupData struct {
	Id int `json:"id"`
}

type AddGroupData struct {
	Name            string `json:"name"`
	SubscriptionUrl string `json:"subscription_url"`
}

type AddProfilesData struct {
	Uris    string `json:"uris"`
	GroupId uint   `json:"group_id"`
}

type DeleteProfilesData struct {
	Profiles []ProfileID `json:"profiles"`
}

type ProfileID struct {
	Id      int `json:"id"`
	GroupId int `json:"group_id"`
}

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:4897")
	if err != nil {
		log.Fatalf("failed to connect %w", err)
	}
	defer conn.Close()

	listen_finished := make(chan bool)

	go listen(conn, listen_finished)

	// send(conn, Message[AddProfilesData]{Msg: "add-profiles", Data: AddProfilesData{
	// 	Uris:    "vless://30f2d443-af46-4dd6-83c9-b5e17299ebd2@104.26.14.69:443?security=tls&sni=carlotta.shoorekeeper.cloudns.org&fp=chrome&type=ws&path=/&host=carlotta.shoorekeeper.cloudns.org&packetEncoding=xudp&encryption=none#[%F0%9F%87%A8%F0%9F%87%A6]t.me/ConfigsHub\n vless://30f2d443-af46-4dd6-83c9-b5e17299ebd2@104.26.14.69:443?security=tls&sni=carlotta.shoorekeeper.cloudns.org&fp=chrome&type=ws&path=/&host=carlotta.shoorekeeper.cloudns.org&packetEncoding=xudp&encryption=none#[%F0%9F%87%A8%F0%9F%87%A6]different0name",
	// 	GroupId: 1}})
	//
	// send(conn, Message[DeleteProfiles]{Msg: "delete-profiles", Data: DeleteProfiles{
	// 	Profiles: []ProfileID{{Id: 10, GroupId: 0}, {Id: 11, GroupId: 0}},
	// }})

	// send(conn, map[string]interface{}{"type": "hello", "value": 123})

	// send(conn, Message[AddGroupData]{Msg: "add-group", Data: AddGroupData{
	// 	Name:            "new_group",
	// 	SubscriptionUrl: "https://none",
	// }})

	// send(conn, Message[DeleteGroupData]{Msg: "delete-group", Data: DeleteGroupData{
	// 	Id: 3,
	// }})

	// send(conn, Message[AddProfilesData]{Msg: "add-profiles", Data: AddProfilesData{
	// 	Uris:    "vless://034175fb-3436-49f3-8ec6-acc1d28b7268@api.dota.website:443?security=tls&sni=api.dota.website&alpn=http/1.1&allowInsecure=1&type=ws&path=/ws/639b01fc6d01f4e9f7e1f19c/D&host=sympathetic.dota.website&packetEncoding=xudp&encryption=none#%F0%9F%9A%80%20@SmoothVPN%20-%20D",
	// 	GroupId: 0}},
	// )

	// send(conn, Message[ConnectData]{Msg: "connect", Data: ConnectData{
	// 	Profile: ProfileID{
	// 		Id:      111,
	// 		GroupId: 0,
	// 	},
	// }})

	// time.Sleep(3 * time.Second)
	// send(conn, Message[DisconnectData]{Msg: "disconnect", Data: DisconnectData{}})
	// for range 20 {
	// 	send(conn, Message[TestProfileData]{Msg: "test-profile", Data: TestProfileData{
	// 		Profile: ProfileID{
	// 			GroupId: 0,
	// 			Id:      2,
	// 		},
	// 	},
	// 	},
	// 	)
	// }
	// send(conn, Message[GetApplicationStateData]{Msg: "get-application-state", Data: GetApplicationStateData{}})

	// send(conn, Message[AddGroupData]{Msg: "add-group", Data: AddGroupData{
	// 	Name:            "new_group",
	// 	SubscriptionUrl: "http://localhost:3949/",
	// }})

	// send(conn, Message[UpdateSubscriptionData]{Msg: "update-subscription",
	// 	Data: UpdateSubscriptionData{
	// 		GroupId: 1,
	// 	},
	// })

	send(conn, Message[DieData]{Msg: "die", Data: DieData{}})

	<-listen_finished
	// select {}
}

func send(conn net.Conn, obj any) {
	data, _ := json.Marshal(obj)
	length := make([]byte, 4)
	binary.BigEndian.PutUint32(length, uint32(len(data)))
	println("len is", len(data))
	conn.Write(length)
	conn.Write(data)
}

func listen(conn net.Conn, listen_finished chan<- bool) {
	defer func() { listen_finished <- true }()
	for {
		lengthBuf := make([]byte, 4)
		reader := bufio.NewReader(conn)

		_, err := io.ReadFull(reader, lengthBuf)

		if err != nil {
			if err != io.EOF {
				log.Printf("Failed to read length , %v", err)
			}
			return
		}

		length := binary.BigEndian.Uint32(lengthBuf)
		if length == 0 || length > 100*1024*1024 {
			log.Printf("Invalid length %d", length)
			return
		}

		payload := make([]byte, length)

		_, err = io.ReadFull(reader, payload)

		if err != nil {
			log.Printf("Failed to read the payload %v", err)
			return
		}

		log.Println(string(payload), "\n\n")

		var raw_tcp_message TcpMessage

		if err := json.Unmarshal(payload, &raw_tcp_message); err != nil {
			log.Printf("Invalid JSON: %v", err)
			return
		}
	}

}
