// Copyright 2025 Sencillo
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

package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
)

var (
	ErrFileFormat = fmt.Errorf("file must be cue, json, or yaml")
)

type cueConfig[T any] struct {
	ctx        *cue.Context
	value      cue.Value
	schema     string
	filePath   string
	userConfig T
}

func Unmarshal[T any](config T, schema, filePath string) (T, error) {
	ext := filepath.Ext(filePath)

	cfg := cueConfig[T]{
		ctx:        cuecontext.New(),
		value:      cue.Value{},
		schema:     schema,
		filePath:   filePath,
		userConfig: config,
	}

	switch ext {
	case ".cue":
		bi := load.Instances([]string{cfg.filePath}, nil)
		cfg.value = cfg.ctx.BuildInstance(bi[0])
	case ".json", ".yaml", ".yml":
		if err := cfg.unmarshalJSON(); err != nil {
			return config, err
		}
	default:
		return config, ErrFileFormat
	}

	return cfg.loadCueConfig()
}

// unmarshalJSON loads data from JSON/YAML files and compiles the cue Value from the data.
func (c *cueConfig[T]) unmarshalJSON() error {
	f, err := os.Open(c.filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	r, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	c.value = c.ctx.CompileBytes(r)

	return nil
}

// loadCueConfig compiles the schema from the cueConfig schema. It then unifies it with the cueConfig value and
// unmarshals that unification into the userConfig object. It returns the userConfig object and an error.
// and then loads that into the config object.
func (c *cueConfig[T]) loadCueConfig() (T, error) {
	s := c.value.Context().CompileString(c.schema)
	u := s.Unify(c.value)
	if err := u.Decode(&c.userConfig); err != nil {
		return c.userConfig, err
	}

	return c.userConfig, nil
}
