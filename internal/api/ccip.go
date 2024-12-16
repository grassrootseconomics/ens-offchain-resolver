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

	var (
		encodedName []byte
		innerData   []byte
	)

	if err := resolveFunc.DecodeArgs(w3.B(r.Data), &encodedName, &innerData); err != nil {
		return err
	}

	ensName := ens.DecodeENSName(encodedName)

	value, err := decodeInnerData(hexutil.Encode(innerData))
	if err != nil {
		return err
	}

	encodedNameHash, err := ens.NameHash(ensName)
	if err != nil {
		return err
	}

	if !bytes.Equal(encodedNameHash[:], value.Bytes()) {
		return ErrNameValidation
	}

	// TODO: Offchain lookup is performed here
	// test.eth -> 0xDeaDbeefdEAdbeefdEadbEEFdeadbeEFdEaDbeeF
	// For now we stub it with a test address

	signature, err := a.ensProvider.SignPayload(
		common.HexToAddress(r.Sender),
		w3.B(r.Data),
		w3.A(testResolvedAddress),
	)
	if err != nil {
		return err
	}

	return httputil.JSON(w, http.StatusOK, OKResponse{
		Ok:          true,
		Description: "CCIP Data",
		Result: map[string]any{
			"decodedName": ensName,
			"nestedData":  hexutil.Encode(innerData),
			"value":       value,
			"sender":      r.Sender,
			"data":        r.Data,
			"sig":         signature,
		},
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
