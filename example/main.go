package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/facebookincubator/zk"
)

func main() {
	path := flag.String("path", "/", "Znode path from which to get data.")
	address := flag.String("server", "127.0.0.1:2181", "Zookeeper server address.")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	client := zk.Client{}
	conn, err := client.DialContext(ctx, "tcp", *address)
	if err != nil {
		fmt.Println("error dialing server:", err)
		return
	}

	data, err := conn.GetData(*path)
	if err != nil {
		fmt.Println("getData error:", err)
		return
	}
	fmt.Printf("Data for node %s: %v\n", *path, string(data))

	children, err := conn.GetChildren(*path)
	if err != nil {
		fmt.Println("getChildren error:", err)
		return
	}

	fmt.Printf("Children of node %s: %v\n", *path, children)
}
