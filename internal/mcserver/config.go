package mcserver

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// Config specifies the config for MCServer
type Config struct {
	Domain    string `json:"domain"`
	Listen    string `json:"listen"`
	Debug     bool   `json:"debug"`
	Ratelimit struct {
		Rate     int `json:"rate"`
		Capacity int `json:"capacity"`
	} `json:"ratelimit"`
}

func readFromFile(fileName string) (*Config, error) {
	if _, err := os.Stat(fileName); err != nil {
		return writeDefaults()
	}

	file, err := os.OpenFile(fileName, os.O_RDONLY, 600)
	if err != nil {
		return nil, fmt.Errorf("opening read-only file: %w", err)
	}

	defer func() {
		_ = file.Close()
	}()

	b, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("reading fileName: %w", err)
	}

	cfg := &Config{}
	if err := json.Unmarshal(b, cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling fileName: %w", err)
	}

	return cfg, nil
}

func writeDefaults() (*Config, error) {
	cfg := &Config{
		Domain: "tunnel.sergitest.dev",
		Listen: ":25565",
		Debug:  false,
		Ratelimit: struct {
			Rate     int `json:"rate"`
			Capacity int `json:"capacity"`
		}{
			150,
			1000,
		},
	}

	b, err := json.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("marshaling default config: %w", err)
	}

	if err := ioutil.WriteFile("config.json", b, 600); err != nil {
		return nil, fmt.Errorf("writing default config file: %w", err)
	}

	return cfg, nil
}
