package api

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/grassrootseconomics/ens-offchain-resolver/pkg/ens"
	"github.com/jackc/pgx/v5"
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
	CELO_COIN = 2147525868

	testResolvedAddress = "0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045"

	AddrSignature      string = "0x3b3b57de"
	MulticoinSignature string = "0xf1cb7e06"
)

var (
	ErrUnsupportedFunction = errors.New("unsupported function")
	ErrNameValidation      = errors.New("could not validate encoded name in inner data")

	resolveFunc = w3.MustNewFunc("resolve(bytes,bytes)", "")

	// https://docs.ens.domains/resolvers/interfaces/#resolver-interface-standards/
	signatures = map[string]*w3.Func{
		AddrSignature:      w3.MustNewFunc("addr(bytes32)", "address"),
		MulticoinSignature: w3.MustNewFunc("addr(bytes32,uint256)", "bytes"),
	}
)

func (a *API) ccipHandler(w http.ResponseWriter, req bunrouter.Request) error {
	w.Header().Set("Access-Control-Allow-Origin", "*")
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

	value, err := a.decodeInnerData(hexutil.Encode(innerData))
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

	address, err := a.store.LookupName(req.Context(), ensName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return httputil.JSON(w, http.StatusBadRequest, CCIPErrResponse{
				Message: "Name not resolved in internal DB.",
			})
		}

		return httputil.JSON(w, http.StatusBadRequest, CCIPErrResponse{
			Message: "Internal server error.",
		})
	}

	resultBytes := a.encodeAddress(hexutil.Encode(innerData), w3.A(address))

	payload, err := a.ensProvider.SignPayload(
		common.HexToAddress(r.Sender),
		w3.B(r.Data),
		resultBytes,
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

func (a *API) decodeInnerData(nestedDataHex string) (*common.Hash, error) {
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
	case MulticoinSignature:
		var (
			result   common.Hash
			coinType *big.Int
		)
		if err := signatures[MulticoinSignature].DecodeArgs(w3.B(nestedDataHex), &result, &coinType); err != nil {
			return nil, err
		}
		a.logg.Debug("decoded coin type", "coinType", coinType)
		if coinType.Cmp(big.NewInt(CELO_COIN)) == 0 {
			return &result, nil
		} else {
			return nil, ErrUnsupportedFunction
		}
	}

	return nil, ErrUnsupportedFunction
}

// TODO: Massive refactor needed here
func (a *API) encodeAddress(nestedDataHex string, addr common.Address) []byte {
	if len(nestedDataHex) < 10 {
		return nil
	}

	switch nestedDataHex[:10] {
	case AddrSignature:
		args := abi.Arguments{
			{Type: abi.Type{T: abi.AddressTy}},
		}

		packedData, err := args.Pack(addr)
		if err != nil {
			panic(err)
		}

		return packedData

	case MulticoinSignature:
		args := abi.Arguments{
			{Type: abi.Type{T: abi.BytesTy}},
		}

		packedData, err := args.Pack(addr.Bytes())
		if err != nil {
			panic(err)
		}

		return packedData
	}
	return nil
}
