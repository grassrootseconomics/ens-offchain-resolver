package api

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/grassrootseconomics/resolver/pkg/ens"
	"github.com/kamikazechaser/common/httputil"
	"github.com/lmittmann/w3"
	"github.com/uptrace/bunrouter"
)

type (
	CCIPParams struct {
		Sender string
		Data   string
	}
)

const (
	testResolvedAddress = "0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045"

	AddrSignature string = "0x3b3b57de"
)

var (
	ErrUnsupportedFunction = errors.New("unsupported function")
	ErrNameValidation      = errors.New("could not validate encoded name in inner data")

	resolveFunc = w3.MustNewFunc("resolve(bytes,bytes)", "")

	signatures = map[string]*w3.Func{
		AddrSignature: w3.MustNewFunc("addr(bytes32)", "address"),
	}
)

func (a *API) ccipHandler(w http.ResponseWriter, req bunrouter.Request) error {
	r := CCIPParams{
		Sender: req.Param("sender"),
	}
	r.Data = strings.TrimSuffix(req.Param("data"), ".json")
	a.logg.Debug("received CCIP request", "sender", r.Sender, "data", r.Data)

	var (
		encodedName []byte
		innerData   []byte
	)

	if err := resolveFunc.DecodeArgs(w3.B(r.Data), &encodedName, &innerData); err != nil {
		return httputil.JSON(w, http.StatusBadRequest, CCIPErrResponse{
			Message: "Could not decode data.",
		})
	}

	ensName := ens.DecodeENSName(encodedName)
	a.logg.Debug("decoded ENS name", "name", ensName)
	a.logg.Debug("decoded inner data", "data", hexutil.Encode(innerData))

	value, err := decodeInnerData(hexutil.Encode(innerData))
	if err != nil {
		if err == ErrUnsupportedFunction {
			return httputil.JSON(w, http.StatusBadRequest, CCIPErrResponse{
				Message: "Unsupported function.",
			})
		}

		return httputil.JSON(w, http.StatusBadRequest, CCIPErrResponse{
			Message: "Bad data.",
		})
	}
	a.logg.Debug("inner data return value", "value", value.Hex())

	encodedNameHash, err := ens.NameHash(ensName)
	if err != nil {
		return httputil.JSON(w, http.StatusBadRequest, CCIPErrResponse{
			Message: "Could not determine encoded name hash.",
		})
	}

	if !bytes.Equal(encodedNameHash[:], value.Bytes()) {
		return httputil.JSON(w, http.StatusBadRequest, CCIPErrResponse{
			Message: "Could not validate name.",
		})
	}

	// TODO: Offchain lookup is performed here
	// *.sarafu.eth -> 0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045
	// For now we stub it with the above test address

	payload, err := a.ensProvider.SignPayload(
		common.HexToAddress(r.Sender),
		w3.B(r.Data),
		w3.A(testResolvedAddress),
	)
	if err != nil {
		return httputil.JSON(w, http.StatusInternalServerError, CCIPErrResponse{
			Message: "Could not sign payload.",
		})
	}
	a.logg.Debug("signed payload", "payload", payload)

	return httputil.JSON(w, http.StatusOK, CCIPOKResponse{
		Data: payload,
	})
}

// For now, we will only support the addr(bytes32) function i.e. Ethereum address only
func decodeInnerData(nestedDataHex string) (*common.Hash, error) {
	if len(nestedDataHex) < 10 {
		return nil, fmt.Errorf("invalid nested data hex")
	}

	switch nestedDataHex[:10] {
	case AddrSignature:
		var result common.Hash

		if err := signatures[AddrSignature].DecodeArgs(w3.B(nestedDataHex), &result); err != nil {
			return nil, err
		}

		return &result, nil
	}

	return nil, ErrUnsupportedFunction
}
