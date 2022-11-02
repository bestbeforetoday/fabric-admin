/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package proposal

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/bestbeforetoday/fabric-admin/internal"
	"github.com/hyperledger/fabric-protos-go-apiv2/common"
)

type transactionContext struct {
	TransactionID   string
	SignatureHeader *common.SignatureHeader
}

func newTransactionContext(signingIdentity *internal.SigningIdentity) (*transactionContext, error) {
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

	transactionCtx := &transactionContext{
		TransactionID:   transactionID,
		SignatureHeader: signatureHeader,
	}
	return transactionCtx, nil
}
