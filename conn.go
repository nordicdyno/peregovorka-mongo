package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
	//"io/ioutil"
	"github.com/apcera/nats"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

type connection struct {
	// The websocket connection
	ws *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	// Names of subscribed channels
	subscribed map[string]bool

	// gnatsd encoded connection
	nec *nats.EncodedConn
	// gnatsd connection
	//nc *nats.Conn
}

// write writes a message with the given message type and payload.
func (c *connection) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

func (c *connection) reader() {
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			break
		}
		c.processMessage(message)
	}
	c.ws.Close()
}

func (c *connection) processMessage(msg []byte) {
	message := SubMessage{}
	err := json.Unmarshal(msg, &message)
	if err != nil {
		log.Println("ERROR: invalid JSON subscribe data" + string(msg))
		return
	}
	log.Printf("message+: %v\n", message)

	s := Subscribe{Id: message.Channel, Conn: c}
	log.Printf("send to channel: %v\n", s)
	h.subscribe <- &s
}

func (c *connection) writer() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			log.Println("... c.write() ... ")
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.write(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			log.Println("send to connection Ping")
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}

	for message := range c.send {
		log.Println("... c.ws.WriteMessage ... ")
		err := c.ws.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			break
		}
	}
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Websocket cookie SID value insert here")

	ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		return
	}

	log.Println("Create ws connection")
	// TODO: Add Id generation
	// FIXME: move magic number to const
	nc, _ := nats.Connect(natsDefaultURL)
	ec, _ := nats.NewEncodedConn(nc, "json")
	c := &connection{
		send:       make(chan []byte, 4096),
		ws:         ws,
		subscribed: make(map[string]bool),
		nec:         ec,
	}
	h.register <- c
	defer func() { h.unregister <- c }()
	go c.writer()
	c.reader()
}
