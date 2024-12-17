package ens

import (
	"crypto/ecdsa"
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

type ENS struct {
	signingKey *ecdsa.PrivateKey
}

const ttl = time.Minute * 5

var eip191Prefix = []byte{0x19, 0x00}

func NewProvider(signingKey *ecdsa.PrivateKey) *ENS {
	return &ENS{
		signingKey: signingKey,
	}
}

func (e *ENS) SignPayload(sender common.Address, request []byte, result common.Address) (string, error) {
	expires := expiryTimestamp()

	payload := encodePayload(sender, expires, request, result)

	sig, err := crypto.Sign(payload.Bytes(), e.signingKey)
	if err != nil {
		return "0x", err
	}
	sig = sig[:64]

	resp, err := encodeABIParameters(common.LeftPadBytes(result.Bytes(), 32), expires, sig)
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

func encodePayload(sender common.Address, expires uint64, request []byte, result common.Address) common.Hash {
	payload := append(eip191Prefix, sender.Bytes()...)
	payload = append(payload, uint64ToBytes(expires)...)
	payload = append(payload, crypto.Keccak256Hash(request).Bytes()...)
	payload = append(payload, crypto.Keccak256Hash(common.LeftPadBytes(result.Bytes(), 32)).Bytes()...)

	return crypto.Keccak256Hash(payload)
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

func uint64ToBytes(value uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, value)
	return b
}
