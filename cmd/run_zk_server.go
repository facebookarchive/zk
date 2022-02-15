/*
 * Copyright (c) Facebook, Inc. and its affiliates.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 *
 */

package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/facebookincubator/zk/integration"
)

func main() {

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	cfg := integration.DefaultConfig()

	server, err := integration.NewZKServer("3.6.2", cfg)
	if err != nil {
		log.Fatalf("unexpected error while initializing zk server: %v", err)
	}

	go func() {
		if err = server.Run(); err != nil {
			log.Fatalf("unexpected error while calling RunZookeeperServer: %s", err)
		}
	}()
	log.Print("Server Started")

	<-done
	log.Print("Server Stopped")

	if err := server.Shutdown(); err != nil {
		log.Fatalf("unexpected error while shutting down zk server: %v", err)
	}
	log.Print("Server Exited Properly")
}
