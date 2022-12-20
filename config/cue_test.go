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

package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

var (
	schema   = `Name: string | *"test"`
	cueData  = `Name: "testing"`
	jsonData = `{"name": "testing"}`
	yamlData = `name: "testing"`
)

type testConfig struct {
	Name string
}

func TestUnmarshal(t *testing.T) {

	tt := []struct {
		name     string
		want     testConfig
		fileName string
		schema   string
		data     string
		err      error
	}{
		{name: "cue file", schema: schema, fileName: "config.cue", data: cueData, want: testConfig{Name: "testing"}, err: nil},
		{name: "json file", schema: schema, fileName: "config.json", data: jsonData, want: testConfig{Name: "testing"}, err: nil},
		{name: "yaml file", schema: schema, fileName: "config.yaml", data: yamlData, want: testConfig{Name: "testing"}, err: nil},
		{name: "file type error", schema: schema, fileName: "config.badext", data: cueData, want: testConfig{Name: "testing"}, err: ErrFileFormat},
	}

	for _, v := range tt {
		t.Run(v.name, func(t *testing.T) {
			dir := t.TempDir()
			fp := filepath.Join(dir, v.fileName)
			if err := os.WriteFile(fp, []byte(v.data), 0644); err != nil {
				t.Fatal(err)
			}

			config, err := Unmarshal(v.want, v.schema, fp)
			if v.err != err {
				t.Errorf("expected error %v but got %v", v.err, err)
			}

			if !reflect.DeepEqual(config, v.want) {
				t.Errorf("expected %v, but got %v", v.want, config)
			}

		})
	}
}
