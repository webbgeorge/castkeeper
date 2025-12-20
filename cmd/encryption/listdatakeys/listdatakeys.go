package listdatakeys

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/webbgeorge/castkeeper/pkg/config/cli"
	"github.com/webbgeorge/castkeeper/pkg/database/encryption"
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
	_, cfg, _, err := cli.ConfigureCLI()
	if err != nil {
		log.Fatal(err)
	}

	dekService, err := encryption.NewDEKService(cfg)
	if err != nil {
		log.Fatal(err)
	}

	keys, err := dekService.ListKeys()
	if err != nil {
		log.Fatal(err)
	}

	// TODO get record count for each key

	// TODO table display
	fmt.Print("ID\tPrimary?\tStatus\n")
	for _, key := range keys {
		fmt.Printf("%d\t%t\t%s\n", key.ID, key.IsPrimary, key.Status)
	}
}
