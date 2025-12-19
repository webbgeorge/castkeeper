package adddatakey

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/webbgeorge/castkeeper/pkg/config/cli"
)

var AddDataKeyCmd = &cobra.Command{
	Use:   "add-data-key",
	Short: "Add a new data encryption key, which becomes the primary",
	Long:  "Adds a new data encryption key to the keyset. The new key is made the primary, otherwise existing keys remain unchanged. Existing data is not re-encrypted.",
	Run:   run,
}

func init() {
	cli.InitGlobalFlags(AddDataKeyCmd)
}

func run(cmd *cobra.Command, args []string) {
	_, _, _, err := cli.ConfigureCLI()
	if err != nil {
		log.Fatal(err)
	}

	// TODO
}
