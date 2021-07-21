package main

import (
	"fmt"
	"net"
	"time"

	"github.com/facebookincubator/zk/flw"
)

func fmain() {
	testFlw()
}

func testFlw() {
	testServers := []string{"localhost:2181"}

	l, err := net.Listen("tcp", fmt.Sprintf("localhost:0"))
	if err != nil {
		fmt.Printf("could not connect: %v\n", err)
		return
	}
	defer l.Close()

	statsSlice, ok := flw.Srvr(testServers, time.Second*10)
	if len(statsSlice) == 0 {
		fmt.Println("no *ServerStats instances returned")
		return
	}
	if !ok {
		fmt.Printf("error getting response for 'srvr': %v\n", statsSlice[0].Error)
		return
	}

	for idx, stat := range statsSlice {
		fmt.Printf("got srvr stat for address %s -> %+v\n", testServers[idx], stat)
	}

	okSlice := flw.Ruok(testServers, time.Second*10)
	if len(okSlice) == 0 {
		fmt.Println("no *ServerStats instances returned")
		return
	}

	for idx, ok := range okSlice {
		fmt.Printf("got ruok response for address %s -> %+v\n", testServers[idx], ok)
	}

	clientsSlice, ok := flw.Cons(testServers, time.Second*10)
	if len(clientsSlice) == 0 || len(clientsSlice[0].Clients) == 0 {
		fmt.Println("no *ServerClient instances returned")
		return
	}

	for idx, clients := range clientsSlice {
		for _, client := range clients.Clients {
			fmt.Printf("got cons client for address %s -> %+v\n", testServers[idx], client)
		}
	}
}
