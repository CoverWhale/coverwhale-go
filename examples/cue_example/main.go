// Copyright 2023 Cover Whale Insurance Solutions Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	_ "embed"
	"fmt"
	"log"

	"github.com/SencilloDev/sencillo-go/config"
)

//go:embed schema.cue
var schema string

type AppConfig struct {
	Port     int
	Address  string
	Protocol string
	Data     Data
}

type Data struct {
	SomeValue string
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
