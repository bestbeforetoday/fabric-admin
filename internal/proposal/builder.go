/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package proposal

import (
	admincommon "github.com/bestbeforetoday/fabric-admin/internal/common"
	"github.com/bestbeforetoday/fabric-admin/pkg/identity"
	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func New(
	signingID identity.SigningIdentity,
	chaincodeName string,
	transactionName string,
	options ...Option,
) (*peer.Proposal, error) {
	transactionCtx, err := newTransactionContext(signingID)
	if err != nil {
		return nil, err
	}

	builder := &builder{
		chaincodeName:   chaincodeName,
		transactionName: transactionName,
		transactionCtx:  transactionCtx,
	}

	if err = admincommon.ApplyOptions(builder, options...); err != nil {
		return nil, err
	}

	return builder.build()
}

type builder struct {
	channelName     string
	chaincodeName   string
	transactionName string
	transactionCtx  *transactionContext
	transient       map[string][]byte
	args            [][]byte
}

func (b *builder) build() (*peer.Proposal, error) {
	headerBytes, err := b.headerBytes()
	if err != nil {
		return nil, err
	}

	chaincodeProposalBytes, err := b.chaincodeProposalPayloadBytes()
	if err != nil {
		return nil, err
	}

	proposal := &peer.Proposal{
		Header:  headerBytes,
		Payload: chaincodeProposalBytes,
	}
	return proposal, nil
}

func (b *builder) headerBytes() ([]byte, error) {
	channelHeaderBytes, err := b.channelHeaderBytes()
	if err != nil {
		return nil, err
	}

	signatureHeaderBytes, err := proto.Marshal(b.transactionCtx.SignatureHeader)
	if err != nil {
		return nil, err
	}

	header := &common.Header{
		ChannelHeader:   channelHeaderBytes,
		SignatureHeader: signatureHeaderBytes,
	}
	return proto.Marshal(header)
}

func (b *builder) channelHeaderBytes() ([]byte, error) {
	extensionBytes, err := proto.Marshal(&peer.ChaincodeHeaderExtension{
		ChaincodeId: &peer.ChaincodeID{
			Name: b.chaincodeName,
		},
	})
	if err != nil {
		return nil, err
	}

	channelHeader := &common.ChannelHeader{
		Type:      int32(common.HeaderType_ENDORSER_TRANSACTION),
		Timestamp: timestamppb.Now(),
		ChannelId: b.channelName,
		TxId:      b.transactionCtx.TransactionID,
		Epoch:     0,
		Extension: extensionBytes,
	}
	return proto.Marshal(channelHeader)
}

func (b *builder) chaincodeProposalPayloadBytes() ([]byte, error) {
	invocationSpecBytes, err := proto.Marshal(&peer.ChaincodeInvocationSpec{
		ChaincodeSpec: &peer.ChaincodeSpec{
			ChaincodeId: &peer.ChaincodeID{
				Name: b.chaincodeName,
			},
			Input: &peer.ChaincodeInput{
				Args: b.chaincodeArgs(),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	chaincodeProposalPayload := &peer.ChaincodeProposalPayload{
		Input:        invocationSpecBytes,
		TransientMap: b.transient,
	}
	return proto.Marshal(chaincodeProposalPayload)
}

func (b *builder) chaincodeArgs() [][]byte {
	result := make([][]byte, len(b.args)+1)

	result[0] = []byte(b.transactionName)
	copy(result[1:], b.args)

	return result
}

// Option implements an option for a transaction proposal.
type Option = func(*builder) error

// WithBytesArguments appends to the transaction function arguments associated with a transaction proposal.
func WithBytesArguments(args ...[]byte) Option {
	return func(b *builder) error {
		b.args = append(b.args, args...)
		return nil
	}
}

// WithArguments appends to the transaction function arguments associated with a transaction proposal.
func WithArguments(args ...string) Option {
	return WithBytesArguments(stringsAsBytes(args)...)
}

func stringsAsBytes(strings []string) [][]byte {
	results := make([][]byte, 0, len(strings))

	for _, v := range strings {
		results = append(results, []byte(v))
	}

	return results
}

// WithTransient specifies the transient data associated with a transaction proposal.
// This is usually used in combination with WithEndorsingOrganizations for private data scenarios
func WithTransient(transient map[string][]byte) Option {
	return func(b *builder) error {
		b.transient = transient
		return nil
	}
}

// WithChannel specifies the name of the channel to which the transaction proposal is directed.
func WithChannel(channelName string) Option {
	return func(b *builder) error {
		b.channelName = channelName
		return nil
	}
}
