package rest

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/gorilla/mux"
	"github.com/icheckteam/ichain/x/identity"
	"github.com/tendermint/go-crypto/keys"
)

type uPass struct {
	LocalAccountName string `json:"account_name"`
	Password         string `json:"password"`
	Sequence         int64
}

func RevokeHandlerFn(cdc *wire.Codec, kb keys.Keybase) func(http.ResponseWriter, *http.Request) {
	ctx := context.NewCoreContextFromViper()
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		var m uPass
		body, err := ioutil.ReadAll(r.Body)
		err = json.Unmarshal(body, &m)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		if m.LocalAccountName == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("account_name is required"))
			return
		}

		if m.Password == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("password is required"))
			return
		}
		info, err := kb.Get(m.LocalAccountName)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
			return
		}
		// build message
		msg := buildRevokeMsg(info.PubKey.Address(), vars["id"])

		// sign
		ctx = ctx.WithSequence(m.Sequence)
		txBytes, err := ctx.SignAndBuild(m.LocalAccountName, m.Password, msg, cdc)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
			return
		}

		// send
		res, err := ctx.BroadcastTx(txBytes)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		output, err := json.MarshalIndent(res, "", "  ")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Write(output)
	}
}

func buildRevokeMsg(creator sdk.Address, claimID string) sdk.Msg {
	return identity.RevokeMsg{
		Sender:  creator,
		ClaimID: claimID,
	}
}