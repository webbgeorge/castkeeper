package deletedatakey

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/webbgeorge/castkeeper/pkg/config/cli"
)

var DeleteDataKeyCmd = &cobra.Command{
	Use:   "delete-data-key",
	Short: "Delete a data encryption key by its ID",
	Run:   run,
}

func init() {
	cli.InitGlobalFlags(DeleteDataKeyCmd)
}

func run(cmd *cobra.Command, args []string) {
	_, _, _, err := cli.ConfigureCLI()
	if err != nil {
		log.Fatal(err)
	}

	// TODO check if key is in use refuse to delete if still in use
	// TODO check if key is disabled
	// TODO confirm deletion - note about DB backups
	// TODO delete the key
}
