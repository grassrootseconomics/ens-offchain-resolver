package api

import (
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/kamikazechaser/common/httputil"
	"github.com/lmittmann/w3"
	"github.com/uptrace/bunrouter"
)

type CCIPParams struct {
	Sender string
	Data   string
}

var (
	resolveFunc = w3.MustNewFunc("resolve(bytes,bytes)", "")

	addrFunc          = w3.MustNewFunc("addr(bytes32)", "")
	addrMulticoinFunc = w3.MustNewFunc("addr(bytes32,uint256)", "")
	textFunc          = w3.MustNewFunc("text(bytes32,string)", "")
	contentHashFunc   = w3.MustNewFunc("contentHash(bytes32)", "")
)

func (a *API) ccipHandler(w http.ResponseWriter, req bunrouter.Request) error {
	r := CCIPParams{
		Sender: req.Param("sender"),
	}
	r.Data = strings.TrimSuffix(req.Param("data"), ".json")

	var (
		encodedName []byte
		data        []byte
	)

	if err := resolveFunc.DecodeArgs(w3.B(r.Data), &encodedName, &data); err != nil {
		return err
	}

	return httputil.JSON(w, http.StatusOK, OKResponse{
		Ok:          true,
		Description: "CCIP Data",
		Result: map[string]any{
			"decodedName": decodeENSName(encodedName),
			"nestedData":  hexutil.Encode(data),
			"sender":      r.Sender,
			"data":        r.Data,
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

func matchSignatureWithFunc(nestedDatHex string) *w3.Func {
	if len(nestedDatHex) < 8 {
		return nil
	}

	switch nestedDatHex[:8] {
	case "0x3b3b57de":
		return addrFunc
	case "0xf1cb7e06":
		return addrMulticoinFunc
	case "0x59d1d43c":
		return textFunc
	case "0xbc1c58d1":
		return contentHashFunc
	}

	return nil
}
