package main

import (
	"fmt"

	"github.com/sergivb01/mctunnel/internal/mcproto"
)

func main() {
	fmt.Printf("CUSTOM_ENDING" + mcproto.CustomEnding)

	fmt.Println(mcproto.ExtractHostPort("mc.hypixel.net." + mcproto.CustomEnding))
	conn := mcproto.NewConnector()
	conn.EncodeDecode()
	//
	// if err := conn.StartAcceptingConnections(context.TODO(), ":25565", 50); err != nil {
	// 	panic(err)
	// }
}
