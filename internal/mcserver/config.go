package mcserver

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// Config specifies the config for MCServer
type Config struct {
	Domain     string   `json:"domain"`
	Listen     string   `json:"listen"`
	Debug      bool     `json:"debug"`
	Production bool     `json:"production"`
	Proxies    []string `json:"proxies"`
	Ratelimit  struct {
		Rate     int `json:"rate"`
		Capacity int `json:"capacity"`
	} `json:"ratelimit"`
}

func readFromFile(fileName string) (*Config, error) {
	if _, err := os.Stat(fileName); err != nil {
		return writeDefaults()
	}

	file, err := os.OpenFile(fileName, os.O_RDONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("opening read-only file: %w", err)
	}

	defer func() {
		_ = file.Close()
	}()

	b, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("reading file data: %w", err)
	}

	cfg := &Config{}
	if err := json.Unmarshal(b, cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling file data: %w", err)
	}

	return cfg, nil
}

func writeDefaults() (*Config, error) {
	cfg := &Config{
		Domain:     "tunnel.sergitest.dev",
		Listen:     ":25565",
		Debug:      false,
		Production: false,
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

	if err := ioutil.WriteFile("config.json", b, 0600); err != nil {
		return nil, fmt.Errorf("writing default config file: %w", err)
	}

	return cfg, nil
}
