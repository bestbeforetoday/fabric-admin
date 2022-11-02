/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package identity

import (
	"github.com/hyperledger/fabric-gateway/pkg/hash"
	gatewayid "github.com/hyperledger/fabric-gateway/pkg/identity"
	"github.com/hyperledger/fabric-protos-go-apiv2/msp"
	"google.golang.org/protobuf/proto"
)

type SigningIdentity interface {
	Creator() ([]byte, error)
	Sign([]byte) ([]byte, error)
}

func NewSigningIdentity(id gatewayid.Identity, sign gatewayid.Sign, hash hash.Hash) SigningIdentity {
	return &signingIdentity{
		id:   id,
		sign: sign,
		hash: hash,
	}
}

type signingIdentity struct {
	id   gatewayid.Identity
	sign gatewayid.Sign
	hash hash.Hash
}

func (signingID *signingIdentity) Creator() ([]byte, error) {
	serializedIdentity := &msp.SerializedIdentity{
		Mspid:   signingID.id.MspID(),
		IdBytes: signingID.id.Credentials(),
	}
	return proto.Marshal(serializedIdentity)
}

func (signingID *signingIdentity) Sign(messageBytes []byte) ([]byte, error) {
	digest := signingID.hash(messageBytes)
	return signingID.sign(digest)
}
