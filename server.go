package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// var (
//
//	PongWaitTime = 1 * time.Minute
//	PingWaitTime = (PongWaitTime * 9) / 10
//
// )
const PongWaitTime = 1 * time.Minute
const PingWaitTime = (PongWaitTime * 9) / 10

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type client struct {
	url        string
	socketConn *websocket.Conn
	handler    func(w http.ResponseWriter, r *http.Request)
}

func NewClient(url string, socket *websocket.Conn) *client {
	handlerFunc := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			msg, err := SerializeRequest(r)
			if err != nil {
				fmt.Println("Unable to serialise request")
			}
			socket.WriteMessage(websocket.BinaryMessage, msg)
			w.WriteHeader(http.StatusAccepted)
		} else {
			w.WriteHeader(http.StatusForbidden)
		}
	}

	return &client{
		url:        url,
		socketConn: socket,
		handler:    handlerFunc,
	}
}

type Manager struct {
	ClientList map[string]client
	sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		ClientList: make(map[string]client),
	}
}

func (s *Manager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	subdomain := strings.Split(r.Host, ".")[0]
	client, ok := s.ClientList[subdomain]
	if ok {
		client.handler(w, r)
	} else {
		w.Write([]byte("client connection closed"))
	}
}

func (m *Manager) AddNewClient(u string, ws *websocket.Conn) {
	m.Lock()
	defer m.Unlock()
	uStruct, _ := url.Parse(u)
	newClient := NewClient(u, ws)
	subdomain := strings.Split(uStruct.Host, ".")[0]
	m.ClientList[subdomain] = *newClient
	go m.HandleClient(newClient)
}

func (m *Manager) RemoveClient(c *client) {
	m.Lock()
	defer m.Unlock()
	clientURL, _ := url.Parse(c.url)
	clientKey := strings.Split(clientURL.Host, ".")[0]
	fmt.Println("\nremoved client : ", c.url)
	c.socketConn.Close()
	delete(m.ClientList, clientKey)
}

func (m *Manager) HandleClient(c *client) {
	c.socketConn.SetReadDeadline(time.Now().Add(PongWaitTime))
	c.socketConn.SetPongHandler(func(appData string) error {
		return c.socketConn.SetReadDeadline(time.Now().Add(PongWaitTime))
	})
	defer m.RemoveClient(c)
	ticker := time.NewTicker(PingWaitTime)
	go func() {
		for {
			<-ticker.C
			c.socketConn.WriteMessage(websocket.PingMessage, []byte(""))
		}
	}()
	for {
		_, _, err := c.socketConn.ReadMessage()
		if err != nil {
			return
		}
	}
}

func NewWebHookHandler(clientsManager *Manager, domain string) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Fatalf("error establishing websocket connection")
		}
		u := []byte(GenerateRandomURL("http", domain, 8))
		clientsManager.AddNewClient(string(u), ws)
		fmt.Printf("\nnew client: %s", u)
		ws.WriteMessage(websocket.TextMessage, u)
	})
	mux.Handle("/", clientsManager)
	return mux
}
