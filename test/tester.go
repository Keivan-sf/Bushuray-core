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

type AddConfig struct {
	Uri     string `json:"uri"`
	GroupId uint   `json:"group_id"`
}

func main() {
	conn, _ := net.Dial("tcp", "127.0.0.1:4897")
	defer conn.Close()

	send(conn, Message[AddConfig]{Msg: "add-config", Data: AddConfig{Uri: "vless://testuri", GroupId: 1}})

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
