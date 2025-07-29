package main

import (
	"bufio"
	"bushuray-core/db"
	"bushuray-core/structs"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

type TcpMessage struct {
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func main() {
	database := db.DB{}
	database.Initialize()
	listen, err := net.Listen("tcp", "127.0.0.1:4897")
	if err != nil {
		log.Fatal(err)
		os.Exit(0)
	}

	fmt.Println("server is listening on port 4897")

	for {
		conn, err := listen.Accept()

		if err != nil {
			log.Printf("failed to accept connection: %v", err)
		}

		go handleConnection(conn, &database)
	}
}

func handleConnection(conn net.Conn, dataBase *db.DB) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		lengthBuf := make([]byte, 4)
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

		var raw_tcp_message TcpMessage

		if err := json.Unmarshal(payload, &raw_tcp_message); err != nil {
			log.Printf("Invalid JSON: %v", err)
			return
		}

		switch raw_tcp_message.Msg {
		case "add-profile":
			var data structs.AddProfileData
			if err := json.Unmarshal(raw_tcp_message.Data, &data); err != nil {
				log.Printf("Invalid body for add-profile %v", err)
				return
			}
			dataBase.AddProfile(data)
			log.Println(data)
		default:
			log.Println("Message not supported")
		}
	}
}
