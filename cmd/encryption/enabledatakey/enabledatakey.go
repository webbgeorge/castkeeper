package enabledatakey

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/webbgeorge/castkeeper/pkg/config/cli"
	"github.com/webbgeorge/castkeeper/pkg/database/encryption"
)

var keyID uint32

var EnableDataKeyCmd = &cobra.Command{
	Use:   "enable-data-key",
	Short: "Enable a disabled data encryption key by its ID",
	Run:   run,
}

func init() {
	cli.InitGlobalFlags(EnableDataKeyCmd)
	EnableDataKeyCmd.Flags().Uint32Var(&keyID, "key-id", 0, "the ID of the data key to enable")
	if err := EnableDataKeyCmd.MarkFlagRequired("key-id"); err != nil {
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

	err = dekService.EnableKey(key.ID)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("key '%d' enabled successfully\n", keyID)
}
