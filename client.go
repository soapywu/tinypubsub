package tinypubsub

import (
	"fmt"

	"github.com/gorilla/websocket"
)

type Client struct {
	*Worker
	ID             ID
	addr           string
	port           int
	messageHandler func(topic Topic, id ID, data []byte)
}

func NewClient(id ID, addr string, port int) (*Client, error) {
	return &Client{
		ID:   id,
		addr: addr,
		port: port,
	}, nil
}

func (c *Client) getConenctUrl() string {
	return fmt.Sprintf("ws://%s:%d%s?%s=%s", c.addr, c.port, WebSocketUrlPath, PeerIDParamKey, c.ID)
}

func (c *Client) Start() error {
	wsUrl := c.getConenctUrl()
	conn, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)
	if err != nil {
		return err
	}
	c.Worker = NewWorker(conn)
	c.Worker.OnMessage(c.processMessage)
	c.Worker.Start()
	Infof("client %s started\n", c.ID)
	return nil
}

func (c *Client) Close() {
	c.Worker.Close()
}

func (c *Client) processMessage(msg []byte) {
	m := DataMessage{}
	if err := m.FromBytes(msg); err != nil {
		Debugln("unexpected msg error", err)
		return
	}
	Debugln("got message", c.ID, m)

	if m.Typ == DataError {
		Debugln("got Error msg", m.Data)
		return
	}

	if c.messageHandler != nil {
		c.messageHandler(m.Topic, m.Publisher, m.Data)
	}
}

func (c *Client) send(msg *ActionMessage) error {
	Debugf("Client %s send msg %v\n", c.ID, msg)
	data, err := msg.ToBytes()
	if err != nil {
		return err
	}
	return c.Worker.Send(data)
}

func (c *Client) Publish(topic Topic, data []byte) error {
	Infof("%s publish topic %s msg\n", c.ID, topic)
	return c.send(&ActionMessage{
		Action: ActionPublish,
		Topic:  topic,
		Data:   data,
	})
}

func (c *Client) Subscribe(topic Topic) error {
	return c.send(&ActionMessage{
		Action: ActionSubscribe,
		Topic:  topic,
	})
}

func (c *Client) Unsubscribe(topic Topic) error {
	return c.send(&ActionMessage{
		Action: ActionUnsubscribe,
		Topic:  topic,
	})
}

func (c *Client) OnMessage(h func(topic Topic, id ID, data []byte)) {
	c.messageHandler = h
}
