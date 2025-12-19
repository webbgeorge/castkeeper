package listdatakeys

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/webbgeorge/castkeeper/pkg/config/cli"
)

var ListDataKeysCmd = &cobra.Command{
	Use:   "list-data-keys",
	Short: "List data encryption keys",
	Run:   run,
}

func init() {
	cli.InitGlobalFlags(ListDataKeysCmd)
}

func run(cmd *cobra.Command, args []string) {
	_, _, _, err := cli.ConfigureCLI()
	if err != nil {
		log.Fatal(err)
	}

	// TODO
}
