package ens

import (
	"crypto/ecdsa"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	solsha3 "github.com/miguelmota/go-solidity-sha3"
)

type ENS struct {
	signingKey *ecdsa.PrivateKey
}

const (
	magicString = "0x1900"
	ttl         = time.Minute * 5
)

func NewProvider(signingKey *ecdsa.PrivateKey) *ENS {
	return &ENS{
		signingKey: signingKey,
	}
}

func (e *ENS) SignPayload(sender common.Address, request []byte, result []byte) (string, error) {
	payload := encodePayload(sender, request, result)

	sig, err := crypto.Sign(payload, e.signingKey)
	if err != nil {
		return "0x", err
	}

	resp, err := encodeABIParameters(payload, expiryTimestamp(), sig)
	if err != nil {
		return "0x", err
	}

	return resp, nil
}

func encodeABIParameters(data []byte, expires uint64, signature []byte) (string, error) {
	arguments := abi.Arguments{
		{Type: abi.Type{T: abi.BytesTy}},
		{Type: abi.Type{T: abi.UintTy, Size: 64}},
		{Type: abi.Type{T: abi.BytesTy}},
	}

	packedData, err := arguments.Pack(data, expires, signature)
	if err != nil {
		return "", fmt.Errorf("failed to encode data: %w", err)
	}

	return hexutil.Encode(packedData), nil
}

func encodePayload(sender common.Address, request []byte, result []byte) []byte {
	return solsha3.SoliditySHA3(
		[]string{"bytes", "address", "uint64", "bytes32", "bytes32"},
		[]interface{}{magicString, sender, expiryTimestamp(), crypto.Keccak256Hash(request).Hex(), crypto.Keccak256Hash(result).Hex()},
	)
}

func expiryTimestamp() uint64 {
	return uint64(time.Now().Add(ttl).Unix())
}

func DecodeENSName(hexBytes []byte) string {
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
