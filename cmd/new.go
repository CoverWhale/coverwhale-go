// Copyright 2025 Sencillo
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"encoding/json"
	"os/exec"

	"github.com/spf13/cobra"
)

// subcommand to create new resources
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Creates a new Sencillo app",
}

func init() {
	rootCmd.AddCommand(newCmd)
}

type Mod struct {
	Path string `json:"Path"`
}

func modInfo() string {
	var mod Mod
	info, err := exec.Command("go", "list", "-json", "-m").Output()
	if err != nil {
		cobra.CheckErr(err)
	}

	if err := json.Unmarshal(info, &mod); err != nil {
		cobra.CheckErr(err)
	}

	return mod.Path
}
