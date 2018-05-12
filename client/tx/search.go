package tx

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	ctypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/wire"
)

const (
	flagTags = "tag"
	flagAny  = "any"
)

// default client command to search through tagged transactions
func SearchTxCmd(cmdr commander) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "txs",
		Short: "Search for all transactions that match the given tags",
		RunE:  cmdr.searchAndPrintTx,
	}
	cmd.Flags().StringP(client.FlagNode, "n", "tcp://localhost:46657", "Node to connect to")
	// TODO: change this to false once proofs built in
	cmd.Flags().Bool(client.FlagTrustNode, true, "Don't verify proofs for responses")
	cmd.Flags().StringSlice(flagTags, nil, "Tags that must match (may provide multiple)")
	cmd.Flags().Bool(flagAny, false, "Return transactions that match ANY tag, rather than ALL")
	return cmd
}

func (c commander) searchTx(tags []string) ([]txInfo, error) {
	if len(tags) == 0 {
		return nil, errors.New("Must declare at least one tag to search")
	}
	// XXX: implement ANY
	query := strings.Join(tags, " AND ")

	// get the node
	node, err := context.NewCoreContextFromViper().GetNode()
	if err != nil {
		return nil, err
	}

	prove := !viper.GetBool(client.FlagTrustNode)
	res, err := node.TxSearch(query, prove)
	if err != nil {
		return nil, err
	}
	return formatTxResults(c.cdc, res)
}

func formatTxResults(cdc *wire.Codec, res []*ctypes.ResultTx) ([]txInfo, error) {
	var err error
	out := make([]txInfo, len(res))
	for i := range res {
		out[i], err = formatTxResult(cdc, res[i])
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

// CMD

func (c commander) searchAndPrintTx(cmd *cobra.Command, args []string) error {
	tags := viper.GetStringSlice(flagTags)
	output, err := c.searchTx(tags)
	if err != nil {
		return err
	}

	fmt.Printf("%+v", output)
	return nil
}

// REST

func SearchTxRequestHandler(cdc *wire.Codec) func(http.ResponseWriter, *http.Request) {
	c := commander{cdc}
	return func(w http.ResponseWriter, r *http.Request) {
		tags := []string{}

		// add hash tag ...
		hash := r.URL.Query().Get("hash")
		if hash != "" {
			tags = append(tags, fmt.Sprintf("tx.hash='%s'", hash))
		}

		// add hash tag ...
		assetID := r.URL.Query().Get("asset_id")
		if assetID != "" {
			tags = append(tags, fmt.Sprintf("asset_id='%s'", assetID))
		}

		// add hash tag ...
		owner := r.URL.Query().Get("owner")
		if owner != "" {
			tags = append(tags, fmt.Sprintf("owner='%s'", owner))
		}

		output, err := c.searchTx(tags)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}

		b, err := json.Marshal(output)
		if err != nil {
			w.WriteHeader(400)
			w.Write([]byte(err.Error()))
			return
		}
		w.Write(b)
	}
}
