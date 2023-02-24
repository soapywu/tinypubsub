package tinypubsub_test

import (
	"testing"
	"time"

	"github.com/soapywu/tinypubsub"
)

func TestMain(t *testing.T) {
	ip := "127.0.0.1"
	port := 9090

	// start server
	server := tinypubsub.NewServer(port)
	server.Start()

	ask := "who am i"
	answer := "you are who you are"
	client1Received := false
	client2Received := false

	errWrapRuner := func(prompt string, f func() error) {
		err := f()
		if err != nil {
			t.Fatal(prompt, err)
		}
	}

	client1, err := tinypubsub.NewClient("client1", ip, port)
	if err != nil {
		t.Fatal("new Client1 failed: ", err)
	}
	errWrapRuner("start Client1 failed: ", func() error {
		return client1.Start()
	})
	client1.OnMessage(func(topic, id tinypubsub.ID, data []byte) {
		if topic != "client1" || id != "client2" || string(data) != answer {
			t.Fatal("Client1 wrong message: ", topic, id, data)
		}
		client1Received = true
	})
	errWrapRuner("Client1 Subscribe failed: ", func() error {
		return client1.Subscribe("client1")
	})

	client2, err := tinypubsub.NewClient("client2", ip, port)
	if err != nil {
		t.Fatal("new Client2 failed: ", err)
	}
	errWrapRuner("start Client2 failed: ", func() error {
		return client2.Start()
	})
	client2.OnMessage(func(topic, id tinypubsub.ID, data []byte) {
		if topic != "client2" || id != "client1" || string(data) != ask {
			t.Fatal("Client2 wrong message: ", topic, id, data)
		}
		client2Received = true

		errWrapRuner("Client1 v failed: ", func() error {
			return client2.Publish("client1", []byte(answer))
		})

	})
	errWrapRuner("Client2 Subscribe failed: ", func() error {
		return client2.Subscribe("client2")
	})

	errWrapRuner("Client1 v failed: ", func() error {
		return client1.Publish("client2", []byte(ask))
	})

	time.Sleep(2 * time.Second)
	if !client1Received || !client2Received {
		t.Fatal("client not received", client1Received, client2Received)
	}
}


