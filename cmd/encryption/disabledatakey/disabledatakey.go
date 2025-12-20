package disabledatakey

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/webbgeorge/castkeeper/pkg/config/cli"
	"github.com/webbgeorge/castkeeper/pkg/database/encryption"
)

var keyID uint32

var DisableDataKeyCmd = &cobra.Command{
	Use:   "disable-data-key",
	Short: "Disable a data encryption key by its ID",
	Run:   run,
}

func init() {
	cli.InitGlobalFlags(DisableDataKeyCmd)
	DisableDataKeyCmd.Flags().Uint32Var(&keyID, "key-id", 0, "the ID of the data key to disable")
	if err := DisableDataKeyCmd.MarkFlagRequired("key-id"); err != nil {
		panic(err)
	}
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

	key, err := dekService.GetKey(keyID)
	if err != nil {
		log.Fatal(err)
	}

	if key.IsPrimary {
		log.Fatal("cannot disable primary key")
	}

	// TODO check that key is not used to encrypt any data

	err = dekService.DisableKey(keyID)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("key '%d' disabled successfully\n", keyID)
}
