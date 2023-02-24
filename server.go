package tinypubsub

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

// constants for server message
const (
	errInvalidActionMessage = "Server: Invalid action msg"
	errActionUnsupported    = "Server: Action unsupported"
	errClientNotExist       = "Server: Not found client"
	errClientIDConflict     = "Server: Client ID conflicted"
	errMessageQueueIsFull   = "Server: Message queue is full"
)

type Connection struct {
	*Worker
	ID     ID
	topics map[Topic]struct{}
}

func (c *Connection) Send(msg *DataMessage) error {
	Debugf("Connection %s send msg %v\n", c.ID, msg)
	data, err := msg.ToBytes()
	if err != nil {
		return err
	}
	return c.Worker.Send(data)
}

func (c *Connection) AddTopic(topic Topic) {
	c.topics[topic] = struct{}{}
}

func (c *Connection) RemoveTopic(topic Topic) {
	delete(c.topics, topic)
}

type ConenctionMap = map[ID]*Connection

type QueueMsg struct {
	Conn *Connection
	Msg  *ActionMessage
}

type Server struct {
	port              int
	subscriptions     map[Topic]ConenctionMap
	connections       ConenctionMap
	websocketUpgrader *websocket.Upgrader
	msgChan           chan *QueueMsg
}

func NewServer(port int) *Server {
	return &Server{
		port:          port,
		subscriptions: make(map[Topic]ConenctionMap),
		connections:   make(ConenctionMap),
		websocketUpgrader: &websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// allow all origin
				return true
			},
		},
		msgChan: make(chan *QueueMsg, 1024),
	}
}

func (s *Server) addConnection(id ID, _conn *websocket.Conn) {
	conn := &Connection{
		Worker: NewWorker(_conn),
		ID:     id,
		topics: make(map[string]struct{}),
	}
	s.connections[id] = conn
	conn.OnClose(func() {
		s.removeConnection(id)
	})
	conn.OnMessage(func(data []byte) {
		s.enqueueMessage(conn, data)
	})
	conn.Start()
	Infof("connection {%s} added\n", id)
}

func (s *Server) removeConnection(id ID) {
	conn := s.connections[id]
	if conn == nil {
		return
	}

	for topic := range conn.topics {
		s.unsubscribe(conn, topic)
	}
	Infof("connection {%s} removed\n", id)
}

func (s *Server) hasConnection(id ID) bool {
	_, found := s.connections[id]
	return found
}

func (s *Server) enqueueMessage(conn *Connection, msg []byte) {
	m := ActionMessage{}
	if err := m.FromBytes(msg); err != nil {
		conn.Send(&DataMessage{Typ: DataError, Data: []byte(errInvalidActionMessage)})
		return
	}

	m.Action = strings.TrimSpace(strings.ToLower(m.Action))
	if !IsValidAction(m.Action) {
		conn.Send(&DataMessage{Typ: DataError, Data: []byte(errActionUnsupported)})
		return
	}

	queueMsg := &QueueMsg{
		Conn: conn,
		Msg:  &m,
	}

	select {
	case s.msgChan <- queueMsg:
	default: // failed to enqueue
		conn.Send(&DataMessage{Typ: DataError, Data: []byte(errMessageQueueIsFull)})
	}
}

func (s *Server) processMessageWorker() {
	for m := range s.msgChan {
		conn := m.Conn
		msg := m.Msg
		switch msg.Action {
		case ActionPublish:
			s.publish(m.Msg.Topic, conn.ID, msg.Data)
		case ActionSubscribe:
			s.subscribe(conn, msg.Topic)
		case ActionUnsubscribe:
			s.unsubscribe(conn, msg.Topic)
		}
	}
}

func (s *Server) publish(topic Topic, publiser ID, data []byte) {
	if _, exist := s.subscriptions[topic]; !exist {
		return
	}

	conns := s.subscriptions[topic]

	var wg sync.WaitGroup
	for id, conn := range conns {
		Debugf("publish topic %s to subscriber %s\n", topic, id)
		GoWithWaitGroup(&wg, func() {
			conn.Send(&DataMessage{
				Typ:       DataMsg,
				Topic:     topic,
				Publisher: publiser,
				Data:      data,
			})
		})
	}

	wg.Wait()
}

func (s *Server) subscribe(conn *Connection, topic Topic) {
	Infof("subscriber %s subscribe topic %s\n", conn.ID, topic)
	conn.AddTopic(topic)
	conns, exist := s.subscriptions[topic]
	if !exist {
		conns = make(ConenctionMap)
		s.subscriptions[topic] = conns
	}

	conns[conn.ID] = conn
}

func (s *Server) unsubscribe(conn *Connection, topic Topic) {
	Infof("subscriber %s unsubscribe topic %s\n", conn.ID, topic)
	conn.RemoveTopic(topic)
	if _, exist := s.subscriptions[topic]; exist {
		delete(s.subscriptions[topic], conn.ID)
		if len(s.subscriptions[topic]) == 0 {
			delete(s.subscriptions, topic)
		}
	}
}

func (s *Server) handleWebsocket(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	peerID := query.Get(PeerIDParamKey)
	if len(peerID) == 0 {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("not peer id supplied"))
		return
	}

	if s.hasConnection(peerID) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("client %s already exist", peerID)))
		return
	}

	conn, err := s.websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed upgrading connection"))
		return
	}
	s.addConnection(peerID, conn)
}

func (s *Server) Start() {
	http.HandleFunc(WebSocketUrlPath, s.handleWebsocket)
	go func() {
		Infoln("start tinypubsub server at port ", s.port)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", s.port), nil); err != nil {
			panic(err)
		}
	}()
	go s.processMessageWorker()
}
