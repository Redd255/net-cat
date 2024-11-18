package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

type server struct {
	listeaAdder string
	Ln          net.Listener
	Clients     map[net.Conn]string
	History     string
}

func newServer(name string) *server {
	return &server{listeaAdder: name}
}

func (s *server) start() {
	ln, err := net.Listen("tcp", s.listeaAdder)
	if err != nil {
		log.Fatal(err)
	}
	s.Ln = ln
	defer ln.Close()
	s.Clients = make(map[net.Conn]string)
	s.acceptConnections()

}

func (s *server) acceptConnections() {
	for {
		conn, err := s.Ln.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go s.readconnection(conn)
	}
}
func (s *server) close(conn net.Conn) {
	conn.Close()
	for client, _ := range s.Clients {
		client.Write([]byte(fmt.Sprintf("%s has left our chat...", strings.TrimSpace(s.Clients[conn]))))
	}
	delete(s.Clients, conn)
}

func messages(conn net.Conn, c chan []byte) {
	buf := make([]byte, 2048)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
		conn.Close()
		return
	}
	msg := buf[:n]
	c <- msg
}

func (s *server) readconnection(conn net.Conn) {
	var mu sync.Mutex
	defer s.close(conn)
	conn.Write([]byte("Welcome to TCP-Chat!\n         _nnnn_\n        dGGGGMMb\n       @p~qp~~qMb\n       M|@||@) M|\n       @,----.JM|\n      JS^\\__/  qKL\n     dZP        qKRb\n    dZP          qKKb\n   fZP            SMMb\n   HZM            MMMM\n   FqM            MMMM\n __| \".        |\\dS\"qML\n |    `.       | `' \\Zq\n_)      \\.___.,|     .'\n\\____   )MMMMMP|   .'\n     `-'       `--'\n[ENTER YOUR NAME]: "))
	buf := make([]byte, 2048)
	n, _ := conn.Read(buf)
	mu.Lock()
	s.Clients[conn] = strings.TrimSpace(string(buf[:n]))
	mu.Unlock()
	conn.Write([]byte(s.History))
	s.History += fmt.Sprintf("%s has joined the chat...\n", s.Clients[conn])
	for client := range s.Clients {
		if client != conn {
			client.Write([]byte(fmt.Sprintf("%s has joined our chat...\n", s.Clients[conn])))
		}
	}
	c := make(chan []byte)

	for {
		go func() {
			messages(conn, c)
		}()
		msg := <-c
		if msg == nil {
			break
		}

		ss := strings.TrimSpace(s.Clients[conn])
		formattedMessage := fmt.Sprintf("[%s][%s]: %s", time.Now().Format("2006-01-02 15:04:05"), ss, msg)
		s.History += formattedMessage

		for client := range s.Clients {
			if client != conn {
				client.Write([]byte(formattedMessage))
			}
		}
	}
}

func main() {
	port := "8080"
	if len(os.Args) == 2 {
		port = os.Args[1]
	}
	s := newServer(":" + port)
	s.start()
}
