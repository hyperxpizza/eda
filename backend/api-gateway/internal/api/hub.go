package api

import (
	"github.com/gorilla/websocket"
	"github.com/hyperxpizza/eda/backend/utils"
	"github.com/sirupsen/logrus"
)

type Hub struct {
	lgr       logrus.FieldLogger
	clients   map[*websocket.Conn]struct{}
	broadcast chan utils.Message
}

func NewHub(lgr logrus.FieldLogger) *Hub {
	return &Hub{
		lgr:       lgr.WithField("module", "hub"),
		clients:   make(map[*websocket.Conn]struct{}),
		broadcast: make(chan utils.Message),
	}
}

func (hub *Hub) Run() {
	hub.lgr.Info("running hub...")
	for {
		select {
		case msg := <-hub.broadcast:
			for client := range hub.clients {
				if err := client.WriteJSON(msg); err != nil {
					hub.lgr.Errorf("failed to write json: %s", err.Error())
				}
			}
		}
	}
}
