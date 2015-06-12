package main

import (
	"github.com/0studio/link"
	"time"
)

// This is broadcast server demo work with the echo_client.
// usage:
//     go run broadcast/main.go
func main() {
	server, err := link.Listen("tcp", "127.0.0.1:10010")
	if err != nil {
		panic(err)
	}

	channel := link.NewChannel(server.Protocol(), link.SERVER_SIDE)
	go func() {
		for {
			time.Sleep(time.Second * 2)
			// broadcast to server sessions
			server.Broadcast(link.String("server say: "+time.Now().String()), 0)
			// broadcast to channel sessions
			channel.Broadcast(link.String("channel say: "+time.Now().String()), 0)
		}
	}()

	println("server start")

	server.Serve(func(session link.SessionAble) {
		println("client", session.Conn().RemoteAddr().String(), "in")
		channel.Join(session, nil)

		session.Process(func(msg *link.InBuffer) error {
			channel.Broadcast(link.String(
				"client "+session.Conn().RemoteAddr().String()+" say: "+string(msg.Data),
			), 0)
			return nil
		})

		println("client", session.Conn().RemoteAddr().String(), "close")
		channel.Exit(session)
	})
}
