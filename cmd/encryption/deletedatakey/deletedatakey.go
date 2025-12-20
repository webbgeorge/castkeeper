package deletedatakey

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/webbgeorge/castkeeper/pkg/cli"
	cliconfig "github.com/webbgeorge/castkeeper/pkg/config/cli"
	"github.com/webbgeorge/castkeeper/pkg/database/encryption"
)

var DeleteDataKeyCmd = &cobra.Command{
	Use:   "delete-data-key",
	Short: "Delete a data encryption key by its ID",
	Run:   run,
}

var keyID uint32

func init() {
	cliconfig.InitGlobalFlags(DeleteDataKeyCmd)
	DeleteDataKeyCmd.Flags().Uint32Var(&keyID, "key-id", 0, "the ID of the data key to delete")
	if err := DeleteDataKeyCmd.MarkFlagRequired("key-id"); err != nil {
		panic(err)
	}
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

	key, err := dekService.GetKey(keyID)
	if err != nil {
		log.Fatal(err)
	}

	if key.IsPrimary {
		log.Fatal("cannot delete primary key")
	}

	if key.Status != encryption.KeyStatusDisabled {
		log.Fatalf("key must be disabled before deleting, current status: '%s'", key.Status)
	}

	// TODO check that key is not used to encrypt any data

	fmt.Println("Deleting a data key is irreversable action, ensure you have backed up the CastKeeper data directory before contiuing.")
	if !cli.Confirm("Are you sure you want to continue?") {
		log.Fatal("aborted")
	}

	err = dekService.DeleteKey(keyID)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("key '%d' deleted successfully\n", keyID)
}
