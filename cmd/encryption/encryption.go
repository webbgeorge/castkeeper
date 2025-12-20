package encryption

import (
	"github.com/spf13/cobra"
	"github.com/webbgeorge/castkeeper/cmd/encryption/adddatakey"
	"github.com/webbgeorge/castkeeper/cmd/encryption/deletedatakey"
	"github.com/webbgeorge/castkeeper/cmd/encryption/disabledatakey"
	"github.com/webbgeorge/castkeeper/cmd/encryption/enabledatakey"
	"github.com/webbgeorge/castkeeper/cmd/encryption/listdatakeys"
	"github.com/webbgeorge/castkeeper/cmd/encryption/reencryptdata"
	"github.com/webbgeorge/castkeeper/cmd/encryption/reencryptdatakeys"
)

var Cmd = &cobra.Command{
	Use:     "encryption",
	Short:   "Encryption key management commands",
	GroupID: "commands",
}

func init() {
	Cmd.AddCommand(adddatakey.AddDataKeyCmd)
	Cmd.AddCommand(deletedatakey.DeleteDataKeyCmd)
	Cmd.AddCommand(disabledatakey.DisableDataKeyCmd)
	Cmd.AddCommand(enabledatakey.EnableDataKeyCmd)
	Cmd.AddCommand(listdatakeys.ListDataKeysCmd)
	Cmd.AddCommand(reencryptdata.ReEncryptDataCmd)
	Cmd.AddCommand(reencryptdatakeys.ReEncryptDataKeysCmd)
}
