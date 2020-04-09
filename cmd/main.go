package main

import (
	"context"
	"fmt"

	"github.com/sergivb01/mctunnel/internal/mcproto"
)

func main() {
	fmt.Println(mcproto.ExtractHostPort("mc.hypixel.net.tunnel.sergitest.dev"))
	if err := mcproto.NewConnector().StartAcceptingConnections(context.TODO(), "localhost:25565", 50); err != nil {
		panic(err)
	}
}
