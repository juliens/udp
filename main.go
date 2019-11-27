package main

import (
	"fmt"
	"log"
	"net"
	"sync"
)

type Client struct {
	L *net.UDPConn
}

var clients = make(map[string]Client)

var clientLock sync.RWMutex

func handle(client net.Addr, data []byte, listener *net.UDPConn) {
	var clientConn Client
	fmt.Println(client)
	clientLock.Lock()
	if _, ok := clients[client.String()]; !ok {
		fmt.Println("NOT FOUND")
		addr2 := "127.0.0.1:8090"
		udpAddr2, err := net.ResolveUDPAddr("udp", addr2)
		if err != nil {
			log.Fatal("Unable to start UDP Server: %s", err)
		}
		listener2, err := net.DialUDP("udp", nil, udpAddr2)

		// fmt.Println("Start Listener")
		if err != nil {
			log.Fatal("Unable to start UDP Server: %s", err)
		}
		// fmt.Println("Listener", listener2)
		clientConn = Client{
			L: listener2,
		}
		fmt.Println("Listener", clientConn.L)

		fmt.Printf("Write %s\n", client)
		clients[client.String()] = clientConn
		fmt.Println(clients)
		go func() {
			fmt.Println("Waiting response")
			bytes := make([]byte, 1024*1024)
			for {
				n, _, _ := listener2.ReadFrom(bytes)

				listener.WriteTo(bytes[:n], client)
			}
		}()
	} else {
		fmt.Println("Already found")
	}
	clientLock.Unlock()
	// fmt.Println("Listener", clientConn.L)

	listenerT := clients[client.String()].L
	listenerT.Write(data)
}

func startUdpEchoServer(addr string) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.Fatal("Unable to start UDP Server: %s", err)
	}
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Fatal("Unable to start UDP Server: %s", err)
	}

	bytes := make([]byte, 1024*1024)

	for {
		// fmt.Println("Waiting")
		n, client, err := udpConn.ReadFrom(bytes)
		// fmt.Println("receive bytes")
		if err != nil {
			log.Fatal("Unable to start UDP Server: %s", err)
		}
		// udpConn.WriteTo(bytes[:n], client)
		handle(client, bytes[:n], udpConn)
	}
}

type Handler interface {
	ServeUDP(*Conn)
}

type Server struct {
	addr string
	handler Handler
}

func (s *Server) Start() {
	udpAddr, err := net.ResolveUDPAddr("udp", s.addr)
	if err != nil {
		log.Fatal("Unable to start UDP Server: %s", err)
	}

	listener, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Fatal("Unable to start UDP Server: %s", err)
	}

	bytes := make([]byte, 1024*1024)
	conns := map[net.Addr]*Conn{}
	for {
		n, client, err := listener.ReadFrom(bytes)
		if err != nil {
			log.Printf("Error while reading: %+v", err)
		}

		var conn *Conn
		if conn, ok := conns[client]; !ok {
			conn = &Conn{client: client, udpConn: listener}
			conns[client] = conn
			if s.handler != nil {
				go func() {
					defer delete(conns, client)
					s.handler.ServeUDP(conn)
				}()
			}
		}
		conn.AddNewBytes(bytes)


	}
}

type Conn struct {
	bytes []byte
	udpConn *net.UDPConn
	client net.Addr
}

func (c *Conn) AddNewBytes(bytes []byte) {

}

func (c *Conn) Read(bytes []byte) (int, error) {
	bytes = c.bytes
	return 0, nil
}

func (c *Conn) Write(bytes []byte) (int, error) {
	return c.udpConn.WriteTo(bytes, c.client)
}

func main() {
	startUdpEchoServer(":8081")
}
