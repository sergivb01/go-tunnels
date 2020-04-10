package proto

type BufferConfig struct {
	HandshakeAddress int `json:"handshake_address"`
}

type Config struct {
	ListenAddress string       `json:"listen_address"`
	MaxPlayers    int          `json:"max_players"`
	Motd          string       `json:"motd"`
	BufferConfig  BufferConfig `json:"buffer_config"`
}

var (
	config = Config{
		ListenAddress: ":25565",
		MaxPlayers:    150,
		Motd:          "Typhoon server",
		BufferConfig: BufferConfig{
			HandshakeAddress: 300,
		},
	}
)
