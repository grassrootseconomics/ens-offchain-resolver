// Copyright 2017 - 2023 Weald Technology Trading.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package namehash provides functions to hash ENS names.
// Minor modifications to error returns have been made to fit the context of this project.

package namehash

import (
	"fmt"
	"strings"

	"golang.org/x/crypto/sha3"
	"golang.org/x/net/idna"
)

var (
	p       = idna.New(idna.MapForLookup(), idna.ValidateLabels(false), idna.CheckHyphens(false), idna.StrictDomainName(false), idna.Transitional(false))
	pStrict = idna.New(idna.MapForLookup(), idna.ValidateLabels(false), idna.CheckHyphens(false), idna.StrictDomainName(true), idna.Transitional(false))
)

// Normalize normalizes a name according to the ENS rules.
func Normalize(input string) (string, error) {
	output, err := p.ToUnicode(input)
	if err != nil {
		return "", fmt.Errorf("failed to convert to standard unicode: %v", err)
	}
	// If the name started with a period then ToUnicode() removes it, but we want to keep it.
	if strings.HasPrefix(input, ".") && !strings.HasPrefix(output, ".") {
		output = "." + output
	}

	return output, nil
}

// LabelHash generates a simple hash for a piece of a name.
func LabelHash(label string) ([32]byte, error) {
	var hash [32]byte

	normalizedLabel, err := Normalize(label)
	if err != nil {
		return [32]byte{}, err
	}

	sha := sha3.NewLegacyKeccak256()
	if _, err = sha.Write([]byte(normalizedLabel)); err != nil {
		return [32]byte{}, fmt.Errorf("failed to write hash: %v", err)
	}
	sha.Sum(hash[:0])

	return hash, nil
}

// NameHash generates a hash from a name that can be used to
// look up the name in ENS.
func NameHash(name string) ([32]byte, error) {
	var hash [32]byte

	if name == "" {
		return hash, nil
	}

	normalizedName, err := Normalize(name)
	if err != nil {
		return [32]byte{}, err
	}
	parts := strings.Split(normalizedName, ".")
	for i := len(parts) - 1; i >= 0; i-- {
		if hash, err = nameHashPart(hash, parts[i]); err != nil {
			return [32]byte{}, err
		}
	}

	return hash, nil
}

func nameHashPart(currentHash [32]byte, name string) ([32]byte, error) {
	var hash [32]byte

	sha := sha3.NewLegacyKeccak256()
	if _, err := sha.Write(currentHash[:]); err != nil {
		return [32]byte{}, fmt.Errorf("failed to write hash: %v", err)
	}
	nameSha := sha3.NewLegacyKeccak256()
	if _, err := nameSha.Write([]byte(name)); err != nil {
		return [32]byte{}, fmt.Errorf("failed to write hash: %v", err)
	}
	nameHash := nameSha.Sum(nil)
	if _, err := sha.Write(nameHash); err != nil {
		return [32]byte{}, fmt.Errorf("failed to write hash: %v", err)
	}
	sha.Sum(hash[:0])

	return hash, nil
}
