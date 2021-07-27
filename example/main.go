package main

import (
	"context"
	"fmt"
	"time"

	"github.com/facebookincubator/zk"
	"github.com/facebookincubator/zk/integration"
)

func main() {
	server, err := integration.NewZKServer("3.6.2", integration.DefaultConfig())
	if err != nil {
		fmt.Println("error initializing server:", err)
		return
	}
	if err = server.Run(); err != nil {
		fmt.Println("error running server:", err)
		return
	}

	defer server.Shutdown()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	client := zk.Client{}
	conn, err := client.DialContext(ctx, "tcp", "127.0.0.1:2181")
	if err != nil {
		fmt.Println("error dialing server:", err)
		return
	}

	data, err := conn.GetData("/")
	if err != nil {
		fmt.Println("getdata error:", err)
		return
	}

	fmt.Println("got data: ", string(data))
}
