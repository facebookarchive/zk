/*
 * Copyright (c) Facebook, Inc. and its affiliates.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 *
 */

package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/facebookincubator/zk"
)

const usage = `
usage: go run example/main.go [-server] [command] [path]

This CLI is used to get data from a given Zookeeper server node.

The first argument represents the command that the user wants to execute. Available commands:
- 'get' - gets data from a Zookeeper node
- 'list' - lists all children of a Zookeeper node

The second argument is used to specify the Zookeeper node path.

Users can specify server address using the --server flag, which is by default set to localhost:2181.

Example usage:
go run example/main.go get /
go run example/main.go --server=192.168.2.3:1234 list /zookeeper
`

func main() {
	flag.Usage = func() {
		if _, err := fmt.Fprint(flag.CommandLine.Output(), usage); err != nil {
			return
		}
	}

	server := flag.String("server", "127.0.0.1:2181", "Zookeeper server address.")
	flag.Parse()

	if len(flag.Args()) != 2 {
		flag.Usage()
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	client := zk.Client{}
	conn, err := client.DialContext(ctx, "tcp", *server)
	if err != nil {
		fmt.Println("error dialing server:", err)
		return
	}

	cmd := flag.Args()[0]
	path := flag.Args()[1]
	switch cmd {
	case "get":
		data, err := conn.GetData(path)
		if err != nil {
			fmt.Println("getData error:", err)
			return
		}
		fmt.Printf("Data for node %s: %v\n", path, string(data))
	case "list":
		children, err := conn.GetChildren(path)
		if err != nil {
			fmt.Println("getChildren error:", err)
			return
		}

		fmt.Printf("Children of node %s: %v\n", path, children)
	default:
		fmt.Printf("Cannot recognize command \"%s\", exiting.\n", cmd)
	}
}
