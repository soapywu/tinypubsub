package tinypubsub

import (
	"github.com/gorilla/websocket"
)

type Worker struct {
	conn           *websocket.Conn
	closeHandler   func()
	messageHandler func(data []byte)
}

func NewWorker(conn *websocket.Conn) *Worker {
	return &Worker{
		conn: conn,
	}
}

func (w *Worker) Start() {
	go w.readLoop()
}

func (w *Worker) Close() {
	w.conn.Close()
	w.onClose()
}

func (w *Worker) OnClose(h func()) {
	w.closeHandler = h
}

func (w *Worker) onClose() {
	if w.closeHandler != nil {
		w.closeHandler()
	}
}

func (w *Worker) OnMessage(h func(data []byte)) {
	w.messageHandler = h
}

func (w *Worker) onMessage(data []byte) {
	if w.messageHandler != nil {
		w.messageHandler(data)
	}
}

func (w *Worker) Send(data []byte) error {
	return w.conn.WriteMessage(websocket.BinaryMessage, data)
}

func (w *Worker) readLoop() {
	for {
		_, msg, err := w.conn.ReadMessage()
		if err != nil {
			w.Close()
			break
		}

		w.onMessage(msg)
	}
}
