/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package identity

import (
	"github.com/hyperledger/fabric-gateway/pkg/hash"
	gatewayid "github.com/hyperledger/fabric-gateway/pkg/identity"
)

type SigningIdentity interface {
	MspID() string
	Credentials() []byte
	Sign([]byte) ([]byte, error)
}

func NewSigningIdentity(id gatewayid.Identity, sign gatewayid.Sign, hash hash.Hash) SigningIdentity {
	return &signingIdentity{
		Identity: id,
		sign:     sign,
		hash:     hash,
	}
}

type signingIdentity struct {
	gatewayid.Identity
	sign gatewayid.Sign
	hash hash.Hash
}

func (signingID *signingIdentity) Sign(messageBytes []byte) ([]byte, error) {
	digest := signingID.hash(messageBytes)
	return signingID.sign(digest)
}
