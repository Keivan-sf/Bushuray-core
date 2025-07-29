package main

import (
	"encoding/binary"
	"encoding/json"
	"net"
)

type Message[T any] struct {
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}

type AddProfiles struct {
	Uris    string `json:"uris"`
	GroupId uint   `json:"group_id"`
}

func main() {
	conn, _ := net.Dial("tcp", "127.0.0.1:4897")
	defer conn.Close()

	send(conn, Message[AddProfiles]{Msg: "add-profiles", Data: AddProfiles{
		Uris:    "vless://30f2d443-af46-4dd6-83c9-b5e17299ebd2@104.26.14.69:443?security=tls&sni=carlotta.shoorekeeper.cloudns.org&fp=chrome&type=ws&path=/&host=carlotta.shoorekeeper.cloudns.org&packetEncoding=xudp&encryption=none#[%F0%9F%87%A8%F0%9F%87%A6]t.me/ConfigsHub\n vless://30f2d443-af46-4dd6-83c9-b5e17299ebd2@104.26.14.69:443?security=tls&sni=carlotta.shoorekeeper.cloudns.org&fp=chrome&type=ws&path=/&host=carlotta.shoorekeeper.cloudns.org&packetEncoding=xudp&encryption=none#[%F0%9F%87%A8%F0%9F%87%A6]different0name",
		GroupId: 0}})

	// send(conn, map[string]interface{}{"type": "hello", "value": 123})
}

func send(conn net.Conn, obj any) {
	data, _ := json.Marshal(obj)
	length := make([]byte, 4)
	binary.BigEndian.PutUint32(length, uint32(len(data)))
	println("len is", len(data))
	conn.Write(length)
	conn.Write(data)
}
