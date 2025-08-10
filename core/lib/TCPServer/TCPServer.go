package TCPServer

import (
	"bufio"
	cmd "bushuray-core/commands"
	"bushuray-core/db"
	appconfig "bushuray-core/lib/AppConfig"
	proxy "bushuray-core/lib/proxy/mainproxy"
	tunmode "bushuray-core/lib/proxy/tun"
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
	clients       map[string]net.Conn
	DB            *db.DB
	mutex         sync.Mutex
	proxy_manager *proxy.ProxyManager
	tun_namager   *tunmode.TunModeManager
	stop_sig      chan<- bool
}

func NewServer(database *db.DB, proxy_manager *proxy.ProxyManager, tun_manager *tunmode.TunModeManager, stop_sig chan<- bool) *Server {
	return &Server{
		DB:            database,
		clients:       make(map[string]net.Conn),
		proxy_manager: proxy_manager,
		tun_namager:   tun_manager,
		stop_sig:      stop_sig,
	}
}

func (s *Server) Start() {
	app_config := appconfig.GetConfig()
	listen, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", app_config.CoreTCPPort))

	if err != nil {
		log.Fatal(err)
		os.Exit(0)
	}

	log.Println("server is listening on port", app_config.CoreTCPPort)

	go s.handleTunModeStatusChange()
	go s.handleStatusChange()
	go s.handleTestResults()

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
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for clientID, conn := range s.clients {

		length := make([]byte, 4)
		binary.BigEndian.PutUint32(length, uint32(len(msg)))

		// log.Println(string(msg))
		_, err := conn.Write(length)
		if err != nil {
			log.Printf("Error sending length %d to %s: %v\n", length, clientID, err)
			conn.Close()
			continue
		}
		_, err = conn.Write(msg)
		if err != nil {
			log.Printf("Error sending %s to $%s: %v\n", msg, clientID, err)
			conn.Close()
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
		case "die":
			s.stop_sig <- true
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

		case "add-group":
			var data structs.AddGroupData
			if err := json.Unmarshal(raw_tcp_message.Data, &data); err != nil {
				log.Printf("Invalid body for add-group %v", err)
				return
			}
			command_handler.AddGroup(data)

		case "delete-group":
			var data structs.DeleteGroupData
			if err := json.Unmarshal(raw_tcp_message.Data, &data); err != nil {
				log.Printf("Invalid body for delete-group %v", err)
				return
			}
			command_handler.DeleteGroup(data)

		case "connect":
			var data structs.ConnectData
			if err := json.Unmarshal(raw_tcp_message.Data, &data); err != nil {
				log.Printf("Invalid body for connect %v", err)
				return
			}
			command_handler.Connect(data, s.proxy_manager, s.tun_namager)

		case "disconnect":
			var data structs.DisconnectData
			if err := json.Unmarshal(raw_tcp_message.Data, &data); err != nil {
				log.Printf("Invalid body for disconnect %v", err)
				return
			}
			command_handler.Disconnect(data, s.proxy_manager, s.tun_namager)

		case "test-profile":
			var data structs.TestProfileData
			if err := json.Unmarshal(raw_tcp_message.Data, &data); err != nil {
				log.Printf("Invalid body for test-profile%v", err)
				return
			}
			go command_handler.TestProfile(data, s.proxy_manager)

		case "get-application-state":
			var data structs.GetApplicationStateData
			if err := json.Unmarshal(raw_tcp_message.Data, &data); err != nil {
				log.Printf("Invalid body for get-application-state%v", err)
				return
			}
			command_handler.GetApplicationState(data, s.proxy_manager, s.tun_namager)

		case "update-subscription":
			var data structs.UpdateSubscriptionData
			if err := json.Unmarshal(raw_tcp_message.Data, &data); err != nil {
				log.Printf("Invalid body for update-subscription%v", err)
				return
			}
			command_handler.UpdateSubscription(data)

		case "enable-tun":
			var data structs.EnableTunData
			if err := json.Unmarshal(raw_tcp_message.Data, &data); err != nil {
				log.Printf("Invalid body for enable-tun%v", err)
				return
			}
			command_handler.EnableTun(data, s.proxy_manager, s.tun_namager)

		case "disable-tun":
			var data structs.DisableTunData
			if err := json.Unmarshal(raw_tcp_message.Data, &data); err != nil {
				log.Printf("Invalid body for disable-tun%v", err)
				return
			}
			command_handler.DisableTun(data, s.tun_namager)

		case "is-root":
			var data structs.IsRootData
			if err := json.Unmarshal(raw_tcp_message.Data, &data); err != nil {
				log.Printf("Invalid body for is-root%v", err)
				return
			}
			command_handler.IsRoot(data)

		default:
			log.Println("Message not supported")
		}
	}
}
