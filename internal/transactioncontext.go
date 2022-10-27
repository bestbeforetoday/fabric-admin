package internal

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/hyperledger/fabric-protos-go-apiv2/common"
)

type TransactionContext struct {
	TransactionID   string
	SignatureHeader *common.SignatureHeader
}

func NewTransactionContext(signingIdentity *SigningIdentity) (*TransactionContext, error) {
	nonce := make([]byte, 24)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	creator, err := signingIdentity.Creator()
	if err != nil {
		return nil, err
	}

	saltedCreator := append(nonce, creator...)
	rawTransactionID := signingIdentity.Hash(saltedCreator)
	transactionID := hex.EncodeToString(rawTransactionID)

	signatureHeader := &common.SignatureHeader{
		Creator: creator,
		Nonce:   nonce,
	}

	transactionCtx := &TransactionContext{
		TransactionID:   transactionID,
		SignatureHeader: signatureHeader,
	}
	return transactionCtx, nil
}
