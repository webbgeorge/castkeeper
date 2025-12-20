package reencryptdata

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/webbgeorge/castkeeper/pkg/config/cli"
)

var ReEncryptDataCmd = &cobra.Command{
	Use:   "re-encrypt-data",
	Short: "Re-encrypt all encrypted data in the database using the current primary data key",
	Run:   run,
}

func init() {
	cli.InitGlobalFlags(ReEncryptDataCmd)
}

func run(cmd *cobra.Command, args []string) {
	_, _, _, err := cli.ConfigureCLI()
	if err != nil {
		log.Fatal(err)
	}

	// TODO
	fmt.Println("NOT IMPLEMENTED")
}
