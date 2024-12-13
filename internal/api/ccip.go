package api

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	namehash "github.com/grassrootseconomics/resolver/pkg"
	"github.com/kamikazechaser/common/httputil"
	"github.com/lmittmann/w3"
	"github.com/uptrace/bunrouter"
)

type CCIPParams struct {
	Sender string
	Data   string
}

const testResolvedAddress = "0xDeaDbeefdEAdbeefdEadbEEFdeadbeEFdEaDbeeF"

var (
	ErrUnsupportedFunction = errors.New("unsupported function")
	ErrNameValidation      = errors.New("could not validate encoded name in inner data")

	resolveFunc = w3.MustNewFunc("resolve(bytes,bytes)", "")
	addrFunc    = w3.MustNewFunc("addr(bytes32)", "address")
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

	result, err := decodeInnerData(hexutil.Encode(innerData))
	if err != nil {
		return err
	}

	encodedNameHash, err := namehash.NameHash(decodeENSName(encodedName))
	if err != nil {
		return err
	}

	if !bytes.Equal(encodedNameHash[:], result.Bytes()) {
		return ErrNameValidation
	}

	// TODO: Offchain lookup is performed here
	// test.eth -> 0xDeaDbeefdEAdbeefdEadbEEFdeadbeEFdEaDbeeF
	// For now we stub it with a test address
	encodedResult := w3.A(testResolvedAddress)

	// EIP-191 signature here

	return httputil.JSON(w, http.StatusOK, OKResponse{
		Ok:          true,
		Description: "CCIP Data",
		Result: map[string]any{
			"decodedName":      decodeENSName(encodedName),
			"nestedData":       hexutil.Encode(innerData),
			"encodedResulthex": encodedResult.Hex(),
			"result":           result,
			"sender":           r.Sender,
			"data":             r.Data,
		},
	})
}

func decodeENSName(hexBytes []byte) string {
	var decodedParts []string
	currentPart := []byte{}

	for _, b := range hexBytes {
		if b == 0 {
			continue
		}
		if b < 32 {
			if len(currentPart) > 0 {
				decodedParts = append(decodedParts, string(currentPart))
				currentPart = []byte{}
			}
		} else {
			currentPart = append(currentPart, b)
		}
	}

	if len(currentPart) > 0 {
		decodedParts = append(decodedParts, string(currentPart))
	}

	return strings.Join(decodedParts, ".")
}

// For now, we will only support the addr(bytes32) function i.e. Ethereum address only
func decodeInnerData(nestedDatHex string) (*common.Hash, error) {
	if len(nestedDatHex) < 10 {
		return nil, fmt.Errorf("invalid nested data hex")
	}

	if nestedDatHex[:10] == "0x3b3b57de" {
		var result common.Hash

		if err := addrFunc.DecodeArgs(w3.B(nestedDatHex), &result); err != nil {
			return nil, err
		}

		return &result, nil
	}

	return nil, ErrUnsupportedFunction
}
