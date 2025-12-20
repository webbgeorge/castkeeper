package reencryptdatakeys

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/webbgeorge/castkeeper/pkg/config/cli"
	"github.com/webbgeorge/castkeeper/pkg/database/encryption"
)

var ReEncryptDataKeysCmd = &cobra.Command{
	Use:   "re-encrypt-data-keys",
	Short: "Re-encrypts the data encryption keyset with a new master key",
	Long:  "Re-encrypts the data encryption keyset with a new master key, depending on the encryption driver. This command can also be used to migrate from one encryption driver to another.",
	Run:   run,
}

func init() {
	cli.InitGlobalFlags(ReEncryptDataKeysCmd)
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

	err = dekService.ReEncryptKeys(newCfg)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("data keys successfully rencrypted with new config")
}
