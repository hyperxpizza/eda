package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/hyperxpizza/eda/backend/utils"
	"github.com/sirupsen/logrus"
)

type Server struct {
	ctx      context.Context
	cancel   context.CancelFunc
	router   *mux.Router
	lgr      logrus.FieldLogger
	hub      *Hub
	producer *Producer
	upgrader websocket.Upgrader
	wg       sync.WaitGroup
}

func NewServer(ctx context.Context, cancel context.CancelFunc, lgr logrus.FieldLogger) *Server {
	srv := Server{
		ctx:    ctx,
		lgr:    lgr.WithField("module", "api-gateway"),
		cancel: cancel,
		wg:     sync.WaitGroup{},
	}

	router := mux.NewRouter()
	router.HandleFunc("/messages", srv.SendMessage).Methods(http.MethodPost)
	router.HandleFunc("/ws", srv.Websocket).Methods(http.MethodGet)

	hub := NewHub(lgr)
	go hub.Run()

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	srv.router = router
	srv.hub = hub
	srv.upgrader = upgrader

	return &srv
}

func (s *Server) WithProducer(p *Producer) *Server {
	s.producer = p
	return s
}

func (s *Server) Run(port string) {
	s.lgr.Infof("running http server...")
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), s.router); err != nil {
		s.lgr.Errorf("serving http failed: %s", err.Error())
		if s.cancel != nil {
			s.cancel()
		}
	}
}

func (s *Server) SendMessage(w http.ResponseWriter, r *http.Request) {
	var message utils.Message
	err := json.NewDecoder(r.Body).Decode(&message)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = s.producer.WriteMessage("messages", message)
	if err != nil {
		s.lgr.Debugf("producer could not produce this message: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	return
}

func (s *Server) Websocket(w http.ResponseWriter, r *http.Request) {
	s.upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer func() {
		delete(s.hub.clients, ws)
		ws.Close()
	}()

	s.hub.clients[ws] = struct{}{}

	for {
		var msg utils.Message
		err := ws.ReadJSON(&msg)
		if err != nil {
			break
		}

		if s.producer != nil {
			s.wg.Add(1)
			go func() {
				if err := s.producer.WriteMessage("messages", msg); err != nil {
					s.lgr.Errorf("write message failed: %s", err.Error())
				}
				s.wg.Done()
			}()
		}

		s.hub.broadcast <- msg
	}

}
