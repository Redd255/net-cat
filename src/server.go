package serverNet

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

const MaxConnections = 2

var c int

type server struct {
	Listenaddress string
	Ln            net.Listener
	Clients       map[net.Conn]string
	History       string
	Mu            sync.Mutex
}

// NewServer initializes a new server
func NewServer(port string) *server {
	return &server{
		Listenaddress: port,
		Clients:       make(map[net.Conn]string),
	}
}

// Start the server and listen for connections
func (s *server) Start() {
	ln, err := net.Listen("tcp", s.Listenaddress)
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	s.Ln = ln
	fmt.Println("Server started on", s.Listenaddress)
	s.AcceptConnections()
}

// Accepts new client connections
func (s *server) AcceptConnections() {
	for {
		c++
		//connection between server and client
		con, err := s.Ln.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}

		// Check if we have reached the maximum number of connections
		if c > MaxConnections {
			fmt.Fprintln(con, "Server is full. Connection rejected.")
			con.Close()
			fmt.Println("Rejected connection from", con.RemoteAddr())
			continue
		}

		// Handle the connection in a new goroutine
		go s.HandleConnection(con)
	}
}

// Manages a single client connection
func (s *server) HandleConnection(con net.Conn) {
	defer s.CloseConnection(con)
	con.Write([]byte("Welcome to TCP-Chat!\n         _nnnn_\n        dGGGGMMb\n       @p~qp~~qMb\n       M|@||@) M|\n       @,----.JM|\n      JS^\\__/  qKL\n     dZP        qKRb\n    dZP          qKKb\n   fZP            SMMb\n   HZM            MMMM\n   FqM            MMMM\n __| \".        |\\dS\"qML\n |    `.       | `' \\Zq\n_)      \\.___.,|     .'\n\\____   )MMMMMP|   .'\n     `-'       `--'\n[ENTER YOUR NAME]: "))
	firstTime := true
naming:
	if !firstTime {
		con.Write([]byte("[INVALIDE NAME RETRY]:"))
	}
	// Reading client's name
	buf := make([]byte, 1024)
	n, err := con.Read(buf)
	if err != nil {
		log.Println("Error reading name:", err)
		return
	}

	clientName := strings.TrimSpace(string(buf[:n]))
	for _, s := range s.Clients {
		if s == clientName {
			firstTime = false
			goto naming
		}
	}
	s.Mu.Lock()
	s.Clients[con] = clientName
	s.Mu.Unlock()
	// Send chat history to the new client
	con.Write([]byte(s.History))

	// Notify other clients
	s.BroadcastMessage(fmt.Sprintf("%s has joined our chat...", clientName), con)

	// Listen for messages from the client
	for {
		buf := make([]byte, 2048)
		con.Write([]byte(fmt.Sprintf("[%s][%s]:", time.Now().Format("2006-01-02 15:04:05"), clientName)))
		n, err := con.Read(buf)
		if err != nil {
			if err.Error() == "EOF" {
				log.Println("Client closed connection gracefully.")
			} else {
				log.Println("Client read error:", err)
			}
			return
		}
		msg := string(buf[:n])
		if strings.TrimSpace(msg) == "" {
			continue
		}
		formattedMessage := fmt.Sprintf("[%s][%s]:%s", time.Now().Format("2006-01-02 15:04:05"), clientName, strings.TrimSpace(string(msg)))

		s.Mu.Lock()
		s.History += formattedMessage + "\n"
		s.Mu.Unlock()

		s.BroadcastMessage(formattedMessage, con)
	}

}

// handles disconnection of a client
func (s *server) CloseConnection(con net.Conn) {
	s.Mu.Lock()
	clientName := s.Clients[con]
	delete(s.Clients, con)
	s.Mu.Unlock()

	// Notify other clients
	exitMessage := fmt.Sprintf("%s has left our chat...", clientName)
	s.BroadcastMessage(exitMessage, nil)
	con.Close()
}

func (s *server) BroadcastMessage(message string, sender net.Conn) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	for client, clientName := range s.Clients {
		if sender != client {
			_, err := client.Write([]byte("\n" + message + "\n" + fmt.Sprintf("[%s][%s]:", time.Now().Format("2006-01-02 15:04:05"), clientName)))
			if err != nil {
				log.Println("Error sending message:", err)
			}
		}
	}
}
