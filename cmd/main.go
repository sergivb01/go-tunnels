package main

import (
	"context"

	"github.com/sergivb01/mctunnel/internal/mcserver"
)

func main() {
	conn := mcserver.NewConnector()
	if err := conn.Start(context.Background(), ":25565", 500); err != nil {
		panic(err)
	}
}
