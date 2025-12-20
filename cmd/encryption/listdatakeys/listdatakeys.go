package listdatakeys

import (
	"log"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/webbgeorge/castkeeper/pkg/cli"
	cliconfig "github.com/webbgeorge/castkeeper/pkg/config/cli"
	"github.com/webbgeorge/castkeeper/pkg/database/encryption"
)

var ListDataKeysCmd = &cobra.Command{
	Use:   "list-data-keys",
	Short: "List data encryption keys",
	Run:   run,
}

func init() {
	cliconfig.InitGlobalFlags(ListDataKeysCmd)
}

func run(cmd *cobra.Command, args []string) {
	_, cfg, _, err := cliconfig.ConfigureCLI()
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

	rows := [][]string{
		{"ID", "Is Primary?", "Status"},
	}
	for _, key := range keys {
		rows = append(rows, []string{
			strconv.FormatInt(int64(key.ID), 10),
			strconv.FormatBool(key.IsPrimary),
			key.Status,
		})
	}
	cli.PrintTable(rows)
}
