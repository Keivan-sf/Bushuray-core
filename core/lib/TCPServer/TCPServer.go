package TCPServer

import (
	"bufio"
	cmd "bushuray-core/commands"
	"bushuray-core/db"
	"bushuray-core/structs"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
)

type Server struct {
	clients map[string]net.Conn
	DB      *db.DB
	mutex   sync.RWMutex
}

func NewServer(database *db.DB) *Server {
	return &Server{
		DB:      database,
		clients: make(map[string]net.Conn),
	}
}

func (s *Server) Start() {
	listen, err := net.Listen("tcp", "127.0.0.1:4897")

	if err != nil {
		log.Fatal(err)
		os.Exit(0)
	}

	fmt.Println("server is listening on port 4897")

	go func() {
		for {
			conn, err := listen.Accept()

			if err != nil {
				log.Printf("failed to accept connection: %v", err)
			}

			s.mutex.Lock()
			clientID := conn.RemoteAddr().String()
			s.clients[clientID] = conn
			s.mutex.Unlock()

			go s.handleConnection(conn, clientID)
		}
	}()
}

func (s *Server) BroadCast(msg []byte) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for clientID, conn := range s.clients {

		length := make([]byte, 4)
		binary.BigEndian.PutUint32(length, uint32(len(msg)))

		log.Println(string(msg))

		_, err := conn.Write(length)
		if err != nil {
			fmt.Printf("Error sending length %d to %s: %v\n", length, clientID, err)
			continue
		}
		_, err = conn.Write(msg)
		if err != nil {
			fmt.Printf("Error sending %s to $%s: %v\n", msg, clientID, err)
			continue
		}
	}
}

func (s *Server) handleConnection(conn net.Conn, clientID string) {
	defer func() {
		conn.Close()
		s.mutex.Lock()
		delete(s.clients, clientID)
		s.mutex.Unlock()
		fmt.Println("Disconnected:", clientID)
	}()

	defer func() {
		fmt.Println("actually disconnected")
	}()

	command_handler := cmd.Cmd{DB: s.DB, Conn: conn, BroadCast: s.BroadCast}
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

		var raw_tcp_message structs.TCPMessage

		if err := json.Unmarshal(payload, &raw_tcp_message); err != nil {
			log.Printf("Invalid JSON: %v", err)
			return
		}

		switch raw_tcp_message.Msg {
		case "add-profiles":
			var data structs.AddProfilesData
			if err := json.Unmarshal(raw_tcp_message.Data, &data); err != nil {
				log.Printf("Invalid body for add-profiles %v", err)
				return
			}
			command_handler.AddProfiles(data)

		case "delete-profiles":
			var data structs.DeleteProfilesData
			if err := json.Unmarshal(raw_tcp_message.Data, &data); err != nil {
				log.Printf("Invalid body for delete-profiles %v", err)
				return
			}
			command_handler.DeleteProfiles(data)

		default:
			log.Println("Message not supported")
		}
	}
}
