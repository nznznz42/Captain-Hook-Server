package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

type clientTestFake struct {
	output map[int][][]byte
	url    string
	domain string
	ws     *websocket.Conn
	client http.Client
}

func (c *clientTestFake) read() {
	for {
		msgType, p, err := c.ws.ReadMessage()
		if err != nil {
			return
		}
		if string(p) != "" {
			if msgType == websocket.TextMessage {
				c.output[websocket.TextMessage] = append(c.output[websocket.TextMessage], p)
			} else if msgType == websocket.BinaryMessage {
				c.output[websocket.BinaryMessage] = append(c.output[websocket.BinaryMessage], p)
			}
		}
	}
}

func NewclientTestFake() *clientTestFake {
	c := &clientTestFake{}
	c.client = http.Client{}
	c.domain = "ws://localhost:8080/ws"
	ws, _, err := websocket.DefaultDialer.Dial(c.domain, nil)
	c.output = make(map[int][][]byte)
	c.ws = ws
	if err != nil {
		log.Fatalf("error in clientTestFake establishing connection with whserver, %v", err.Error())
	}

	for {
		_, p, err := c.ws.ReadMessage()
		if err != nil {
			continue
		}
		if string(p) != "" {
			c.url = string(p)
			break
		}
	}
	c.ws = ws
	return c
}

type webhookTrigger struct {
	url string
}

func (w webhookTrigger) sendPOSTRequest() {
	body := bytes.NewReader([]byte("hello world"))
	_, err := http.Post(w.url, "applications/json", body)
	if err != nil {
		log.Fatalf("error making post request to the server, %v", err.Error())
	}
}

func (w webhookTrigger) sentGETRequest() *http.Response {
	resp, err := http.Get(w.url)
	if err != nil {
		log.Fatalf("error making get request to the server, %v", err.Error())
	}
	return resp
}

func TestRandomURL(t *testing.T) {
	t.Run("genereates a URL with 8 character length subdomain", func(t *testing.T) {
		rawURL := GenerateRandomURL("http", "localhost:8080", 8)
		paresedurl, err := url.Parse(rawURL)
		if err != nil {
			t.Errorf("unable to parse rawURL, got %s, error : %v", rawURL, err)
		}
		subdomain := strings.Split(paresedurl.Host, ".")[0]
		want := 8
		if len(subdomain) != want {
			t.Errorf("invlaid subdomain length, got %d(%s), want %d", len(subdomain), subdomain, want)
		}
	})

	t.Run("generate a valid URL", func(t *testing.T) {
		url := GenerateRandomURL("http", "localhost:8080", 8)
		if !CheckValidURL(url) {
			t.Errorf("generates an invalid url got %s", url)
		}
	})
	t.Run("generates a random URL everytime", func(t *testing.T) {
		var urlList []string
		for i := 0; i < 10; i++ {
			url := GenerateRandomURL("http", "localhost:8080", 8)
			for _, u := range urlList {
				if u == url {
					t.Fatalf("found duplicate urls, %s", u)
				}
			}
			urlList = append(urlList, url)
		}
	})
}

func TestForwardingMessage(t *testing.T) {
	close := startServer(t)
	defer close()
	time.Sleep(1 * time.Second)

	t.Run("server pings the client on POST request to temp URL", func(t *testing.T) {

		c := NewclientTestFake()
		defer c.ws.Close()

		go func() {
			c.read()
		}()

		whTrigger := webhookTrigger{
			url: c.url,
		}
		whTrigger.sendPOSTRequest()

		if len(c.output) == 0 {
			t.Error("expected POST request to be forwarded by the server, but got none")
		}
	})

	t.Run("server forwards only post request", func(t *testing.T) {
		c := NewclientTestFake()
		defer c.ws.Close()

		go func() {
			c.read()
		}()

		whTrigger := webhookTrigger{
			url: c.url,
		}
		resp := whTrigger.sentGETRequest()
		if resp.StatusCode == http.StatusAccepted {
			t.Errorf("get requets not ignored got statusAccepted, %d", resp.StatusCode)
		}
	})
	t.Run("server sends the post request it receives in binary format to client", func(t *testing.T) {
		c := NewclientTestFake()
		defer c.ws.Close()

		go func() {
			c.read()
		}()

		data := []byte("this is the body of the new post request")
		body := bytes.NewBuffer(data)
		req, err := http.NewRequest(http.MethodPost, c.url, body)
		if err != nil {
			t.Errorf("error creating request, %v", err)
		}

		c.client.Do(req)

		if len(c.output[websocket.BinaryMessage]) == 0 {
			t.Fatalf("didn't receive any binary message")
		}
	})
}

func TestClientConnection(t *testing.T) {
	manager := NewManager()
	mux := NewWebHookHandler(manager, "localhost:8080")
	srv := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	go srv.ListenAndServe()
	defer srv.Close()

	t.Run("parallel test group", func(t *testing.T) {
		t.Run("server listens on generated URL", func(t *testing.T) {
			t.Parallel()
			c := NewclientTestFake()
			defer c.ws.Close()

			u := c.url
			res, err := http.Head(u)
			if err != nil {
				t.Errorf("error making http HEAD, %s", err.Error())
			}
			if res.StatusCode == 404 {
				t.Errorf("server is not listenting on URL, %s", u)
			}
		})

		t.Run("server sends ping messages to the client", func(t *testing.T) {
			t.Parallel()
			received := make(chan bool)
			c := NewclientTestFake()
			c.ws.SetPingHandler(func(appData string) error {
				received <- true
				return nil
			})
			go c.read()
			select {
			case <-received:
				return
			case <-time.After(2 * time.Second):
				t.Fatal("didn't receive any pong messages from server")
			}
		})

		t.Run("disconnected client connection is deleted from Manager", func(t *testing.T) {
			t.Parallel()
			c := NewclientTestFake()
			c.ws.SetPingHandler(func(appData string) error {
				return nil
			})

			time.Sleep((PongWaitTime * 13) / 10)

			u, _ := url.Parse(c.url)
			_, ok := manager.ClientList[u.Host]
			if ok == true {
				t.Fatal("client is present in manager")
			}

			checkClientConnectionClose(t, manager, c)

		})

		t.Run("server removes client if, client closes websocket connection", func(t *testing.T) {
			t.Parallel()

			c := NewclientTestFake()
			c.ws.Close()

			time.Sleep(PongWaitTime)

			checkClientConnectionClose(t, manager, c)
		})

		t.Run("clients which pong's server, maintains connection", func(t *testing.T) {
			t.Parallel()
			c := NewclientTestFake()
			go c.read()

			time.Sleep(PongWaitTime)

			checkClientConnectionOpen(t, manager, c)
		})

		t.Run("server removes client if client closes websocket connection", func(t *testing.T) {
			t.Parallel()
			c := NewclientTestFake()
			c.ws.Close()
			time.Sleep(PongWaitTime)
			checkClientConnectionClose(t, manager, c)
		})
	})
}

func checkClientConnectionClose(t testing.TB, manager *Manager, c *clientTestFake) {
	u, _ := url.Parse(c.url)
	subdomain := strings.Split(u.Host, ".")[0]
	_, ok := manager.ClientList[subdomain]
	if ok == true {
		t.Fatal("client is present in manager")
	}

	var err error
	var closed = make(chan bool)
	go func() {
		_, _, err = c.ws.ReadMessage()
		if err != nil {
			closed <- true
		}
	}()

	select {
	case <-closed:
		return
	case <-time.After(3 * time.Second):
		t.Fatal("websocket connection is not closed")
	}
}

func checkClientConnectionOpen(t testing.TB, manager *Manager, c *clientTestFake) {
	u, _ := url.Parse(c.url)
	subdomain := strings.Split(u.Host, ".")[0]
	_, ok := manager.ClientList[subdomain]
	if ok == false {
		t.Fatal("client is not present in manager")
	}

	var err error
	var closed = make(chan bool)
	go func() {
		_, _, err = c.ws.ReadMessage()
		if err != nil {
			closed <- true
		}
	}()

	select {
	case <-closed:
		t.Fatalf("client connection is closed")
	case <-time.After(3 * time.Second):
		return
	}
}

func startServer(t testing.TB) func() error {
	t.Helper()
	clientManager := NewManager()
	mux := NewWebHookHandler(clientManager, "localhost:8080")
	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	go func() {
		fmt.Println(srv.ListenAndServe())
	}()
	return srv.Close
}

func CheckValidURL(u string) bool {
	ustruct, err := url.ParseRequestURI(u)
	if err != nil {
		return false
	}
	if !(ustruct.Scheme == "http" || ustruct.Scheme == "https") || (ustruct.Host == "") {
		return false
	}
	return true
}
