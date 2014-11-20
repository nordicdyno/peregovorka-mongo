package main

import (
	//"github.com/apcera/nats"
	"github.com/davecgh/go-spew/spew"

	"log"
)

// explicit url value
const (
	natsDefaultURL = "nats://localhost:4222"
)

type Subscribe struct {
	Id   string
	Conn *connection
}

type SubMessage struct {
	Event   string
	Channel string `json:"channel"`
	// Data    SubMessageData `json:"data"` // just ignore it
}

type hub struct {
	// Registered connections.
	connections map[*connection]bool

	// register requests from the connections.
	register chan *connection

	// Unregister requests from connections.
	unregister chan *connection

	// Register Nats Connection
	subscribe chan *Subscribe
}

var h = hub{
	connections: make(map[*connection]bool),
	register:    make(chan *connection),
	unregister:  make(chan *connection),
	subscribe:   make(chan *Subscribe),
}

func (h *hub) run() {
	log.Printf("hub start of life\n")
	for {
		select {
		case c := <-h.register:
			h.connections[c] = true
		case c := <-h.unregister:
			log.Println("h.unregister fired")

			delete(h.connections, c)
			log.Println("connection deleted from hub's poll OK")

			/*
				for topic, reader := range c.subscribed {
					log.Println("delete connection on topic:" + topic)
					reader.RemoveConnection(c)
				}
			*/
			close(c.send)
			//c.subscribed = nil
		case sub := <-h.subscribe:
			log.Println("h.addSubscriber fired")
			//_ = sub
			c * connection
			c = sub.Conn
			_, ok := h.connections[c]
			if !ok {
				log.Println("ERROR: connection for subscribe not found '%s'", sub.Id)
				break
			}
			log.Println("Connection found '%v'", c)
			c.nec.Subscribe(sub.Id, subHandle)
			// c.nats.
			//c.
			/*
				nsqTopicName := GenNSQtopicName(sub.Topic)
				reader, is_reader_exists := h.nsqReaders[nsqTopicName]

				c := sub.Conn
				if !is_reader_exists {
					var err error
					reader, err = NewNsqTopicReader(sub.Topic)
					if err != nil {
						log.Printf("failed to subscribe to topic '%s'", sub.Topic)
						break
					}
				}
				c.subscribed[nsqTopicName] = reader
				h.nsqReaders[nsqTopicName] = reader

				reader.AddConnection(c, sub.Topic)
			*/
		}
	}

	panic("hub end of life")
}

type GenericMessage map[string]interface{}

func subHandle(m *PostMessage) {
	log.Println("sub recieved message", spew.Sdump(m))
}
