package main

import (
	_ "embed"
	"fmt"
	"log"

	"github.com/CoverWhale/coverwhale-go/config"
)

//go:embed schema.cue
var schema string

type AppConfig struct {
	Port     int
	Address  string
	Protocol string
}

func main() {
	cfg, err := config.Unmarshal(AppConfig{}, schema, "./config.cue")
	if err != nil {
		log.Fatal(err)
	}

	jsonCfg, err := config.Unmarshal(AppConfig{}, schema, "./config.json")
	if err != nil {
		log.Fatal(err)
	}

	yamlCfg, err := config.Unmarshal(AppConfig{}, schema, "./config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(cfg)
	fmt.Println(jsonCfg)
	fmt.Println(yamlCfg)
}
