package version

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/webbgeorge/castkeeper"
)

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the CastKeeper version",
	Long:  "Show the CastKeeper version.",
	Run:   run,
}

func run(cmd *cobra.Command, args []string) {
	fmt.Printf("CastKeeper version %s\n", castkeeper.Version)
}
