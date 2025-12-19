package disabledatakey

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/webbgeorge/castkeeper/pkg/config/cli"
)

var DisableDataKeyCmd = &cobra.Command{
	Use:   "disable-data-key",
	Short: "Disable a data encryption key by its ID",
	Run:   run,
}

func init() {
	cli.InitGlobalFlags(DisableDataKeyCmd)
}

func run(cmd *cobra.Command, args []string) {
	_, _, _, err := cli.ConfigureCLI()
	if err != nil {
		log.Fatal(err)
	}

	// TODO
}
