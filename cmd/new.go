package cmd

import (
	"encoding/json"
	"os/exec"

	"github.com/spf13/cobra"
)

// subcommand to create new resources
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Creates a new type of Cover Whale app",
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
