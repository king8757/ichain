package commands

import (
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/client/builder"
	"github.com/cosmos/cosmos-sdk/client/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	wire "github.com/cosmos/cosmos-sdk/wire"
	"github.com/icheckteam/ichain/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// GetCreateAssetAccountTxCmd returns a createAssetAccountTxCmd.
func GetCreateAssetAccountTxCmd(cdc *wire.Codec) *cobra.Command {
	cmdr := Commander{Cdc: cdc}
	cmd := &cobra.Command{
		Use:   "create-asset",
		Short: "Create and sign a CreateAssetAccountTx",
		RunE:  cmdr.createAssetAccountTxCmd,
		Args:  cobra.ExactArgs(1),
	}
	cmd.Flags().String(flagPubKey, "", "New assset account's pubkey")
	cmd.Flags().Int64(flagSequence, 0, "Sequence number")
	return cmd
}

func (c Commander) createAssetAccountTxCmd(cmd *cobra.Command, args []string) error {
	name := args[0]
	keybase, err := keys.GetKeyBase()
	if err != nil {
		return nil
	}
	info, err := keybase.Get(name)
	if err != nil {
		return err
	}
	creator := info.PubKey.Address()
	msg, err := buildCreateAssetAccountMsg(creator)
	if err != nil {
		return err
	}

	res, err := builder.SignBuildBroadcast(name, msg, c.Cdc)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Committed at block %d. Hash: %s\n", res.Height, res.Hash.String())
	return nil

}

func buildCreateAssetAccountMsg(creator sdk.Address) (sdk.Msg, error) {
	// parse new account pubkey
	pubKey, err := types.PubKeyFromHexString(viper.GetString(flagPubKey))
	if err != nil {
		return nil, err
	}
	msg := types.NewCreateAssetAccountMsg(creator, pubKey)
	return msg, nil
}
