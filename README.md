# tinypubsub

A tiny Pub/Sub tool over websockets

## Example

### Server

```go
import (
    "github.com/soapywu/tinypubsub"
)

port := 9090
server := tinypubsub.NewServer(port)
server.Start()
```

### Client

```go
import (
    "fmt"
    "github.com/soapywu/tinypubsub"
)

serverIp := "127.0.0.1"
serverPort := 9090
topic := "test"
id := "client"
msg := "who am i"

client, _ := tinypubsub.NewClient(id, serverIp, serverPort)
client.Start()
client.OnMessage(func(topic, id tinypubsub.ID, data []byte) {
    fmt.Printf("recv topic %s message %s from %s", topic, string(data), id)
})
client.Subscribe(topic)
client.Publish(topic, []byte(msg))
```
