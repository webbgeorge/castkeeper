package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/webbgeorge/castkeeper/cmd/createuser"
	"github.com/webbgeorge/castkeeper/cmd/serve"
	"github.com/webbgeorge/castkeeper/cmd/version"
)

func main() {
	userRootCmd := &cobra.Command{Use: "user"}
	userRootCmd.AddCommand(createuser.CreateUserCmd)

	rootCmd := &cobra.Command{Use: "castkeeper"}
	rootCmd.AddCommand(
		serve.ServeCmd,
		userRootCmd,
		version.VersionCmd,
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
