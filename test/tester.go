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

	// send(conn, Message[AddProfiles]{Msg: "add-profiles", Data: AddProfiles{
	// 	Uris:    "vless://30f2d443-af46-4dd6-83c9-b5e17299ebd2@104.26.14.69:443?security=tls&sni=carlotta.shoorekeeper.cloudns.org&fp=chrome&type=ws&path=/&host=carlotta.shoorekeeper.cloudns.org&packetEncoding=xudp&encryption=none#[%F0%9F%87%A8%F0%9F%87%A6]t.me/ConfigsHub\n vless://30f2d443-af46-4dd6-83c9-b5e17299ebd2@104.26.14.69:443?security=tls&sni=carlotta.shoorekeeper.cloudns.org&fp=chrome&type=ws&path=/&host=carlotta.shoorekeeper.cloudns.org&packetEncoding=xudp&encryption=none#[%F0%9F%87%A8%F0%9F%87%A6]different0name",
	// 	GroupId: 0}})

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


	send(conn, Message[ConnectData]{Msg: "connect", Data: ConnectData{
		Profile: ProfileID{
			Id:      1,
			GroupId: 0,
		},
	}})

	// <-listen_finished
	select {}
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

		log.Println(string(payload))

		var raw_tcp_message TcpMessage

		if err := json.Unmarshal(payload, &raw_tcp_message); err != nil {
			log.Printf("Invalid JSON: %v", err)
			return
		}
	}

	listen_finished <- true
}
