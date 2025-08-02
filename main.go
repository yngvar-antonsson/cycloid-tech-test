package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"

	"cycloid/test/handler"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Providers map[string]map[string]any `yaml:"providers"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return &cfg, nil
}

func main() {
	port := flag.Int("port", 8080, "Port for the server")
	apiLimit := flag.Int("api-limit", 30, "Limit for api calls in seconds")
	configPath := flag.String("config", "config.yml", "Path to configuration file")
	flag.Parse()

	if *port < 0 || *port >= int(math.Pow(2, 16)) {
		log.Fatal("invalid port number")
	}

	if *apiLimit <= 0 {
		log.Fatal("invalid API calls limit")
	}

	config, err := LoadConfig(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	handler.Setup(*apiLimit, config.Providers)

	http.HandleFunc("/weather", handler.WeatherHandler)
	log.Printf("Server started at :%d", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
