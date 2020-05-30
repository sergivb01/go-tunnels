package main

import (
	"context"
	"flag"

	"github.com/minebreach/go-tunnels/internal/mcserver"
)

var fileFlag = flag.String("config", "config.json", "Configuration file path - config.json")

func main() {
	flag.Parse()

	conn, err := mcserver.NewConnector(*fileFlag)
	if err != nil {
		panic(err)
	}

	if err := conn.Start(context.Background()); err != nil {
		panic(err)
	}
}
