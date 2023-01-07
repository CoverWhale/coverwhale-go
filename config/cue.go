package config

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"cuelang.org/go/encoding/json"
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
		if err := cfg.unmarshalCue(); err != nil {
			return config, err
		}
	case ".json", ".yaml", ".yml":
		if err := cfg.unmarshalJSON(); err != nil {
			return config, err
		}
	default:
		return config, ErrFileFormat
	}

	return cfg.loadCueConfig()
}

func (c *cueConfig[T]) thing(f string) {
	r, err := os.Open(f)
	if err != nil {
		log.Fatal(err)
	}
	b, err := json.NewDecoder(nil, c.filePath, r).Extract()
	if err != nil {
		log.Println(err)
	}

	val := c.ctx.BuildExpr(b, nil)
	fmt.Println(val)
}

// unmarshalCue loads the cue file from the cueConfig filePath. It builds the package instances and sets the
// cueConfig value to the first instance.
func (c *cueConfig[T]) unmarshalCue() error {
	buildinstances := load.Instances([]string{c.filePath}, nil)

	insts, err := c.ctx.BuildInstances(buildinstances)
	if err != nil {
		return err
	}

	c.value = insts[0].Value()

	return nil
}

// unmarshalJSON is similar to the unmarshalCue() method but reads JSON/YAML files. It loads the data from the
// cueConfig filePath and compiles the cue Value from the data.
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
