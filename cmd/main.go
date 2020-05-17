package main

import (
	"context"

	"github.com/minebreach/go-tunnels/internal/mcserver"
)

func main() {
	conn := mcserver.NewConnector()
	if err := conn.Start(context.Background(), ":25565", 500); err != nil {
		panic(err)
	}
}
