package internal

import (
	"errors"

	"github.com/hyperledger/fabric-gateway/pkg/hash"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"github.com/hyperledger/fabric-protos-go-apiv2/msp"
	"google.golang.org/protobuf/proto"
)

type SigningIdentity struct {
	ID   identity.Identity
	Sign identity.Sign
	Hash hash.Hash
}

func NewSigningIdentity(id identity.Identity) *SigningIdentity {
	return &SigningIdentity{
		ID: id,
		Sign: func(digest []byte) ([]byte, error) {
			return nil, errors.New("no sign implementation supplied")
		},
		Hash: hash.SHA256,
	}
}

func (signingID *SigningIdentity) Creator() ([]byte, error) {
	serializedIdentity := &msp.SerializedIdentity{
		Mspid:   signingID.ID.MspID(),
		IdBytes: signingID.ID.Credentials(),
	}
	return proto.Marshal(serializedIdentity)
}
