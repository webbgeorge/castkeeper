package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/webbgeorge/castkeeper/cmd/encryption"
	"github.com/webbgeorge/castkeeper/cmd/serve"
	"github.com/webbgeorge/castkeeper/cmd/users"
	"github.com/webbgeorge/castkeeper/cmd/version"
)

func main() {
	rootCmd := &cobra.Command{Use: "castkeeper"}

	rootCmd.AddGroup(&cobra.Group{
		ID:    "commands",
		Title: "Commands:",
	})

	rootCmd.AddCommand(
		serve.ServeCmd,
		users.Cmd,
		encryption.Cmd,
		version.VersionCmd,
	)

	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "enable verbose output")
	rootCmd.PersistentFlags().StringP("config", "c", "", "config file (otherwise uses default locations)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
