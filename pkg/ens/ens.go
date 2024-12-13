package ens

import (
	"crypto/ecdsa"
	"time"

	"github.com/ethereum/go-ethereum/common"
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

func (e *ENS) SignPayload(sender common.Address, request []byte, result []byte) ([]byte, error) {
	sig, err := crypto.Sign(encodePayload(sender, request, result), e.signingKey)
	if err != nil {
		return nil, err
	}

	return sig, nil
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
