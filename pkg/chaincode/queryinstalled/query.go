/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package queryinstalled

import (
	"context"
	"errors"
	"fmt"

	"github.com/bestbeforetoday/fabric-admin/internal/common"
	"github.com/bestbeforetoday/fabric-admin/internal/proposal"
	"github.com/bestbeforetoday/fabric-admin/pkg/identity"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer/lifecycle"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

const queryInstalledTransactionName = "QueryInstalledChaincodes"

func Query(ctx context.Context, signingID identity.SigningIdentity, options ...Option) (*lifecycle.QueryInstalledChaincodesResult, error) {
	installCommand := &command{
		signingID: signingID,
	}

	if err := common.ApplyOptions(installCommand, options...); err != nil {
		return nil, err
	}

	return installCommand.run(ctx)
}

type command struct {
	signingID   identity.SigningIdentity
	grpcClient  peer.EndorserClient
	grpcOptions []grpc.CallOption
}

func (c *command) run(ctx context.Context) (*lifecycle.QueryInstalledChaincodesResult, error) {
	if err := c.validate(); err != nil {
		return nil, err
	}

	signedProposal, err := c.signedProposal()
	if err != nil {
		return nil, err
	}

	proposalResponse, err := c.grpcClient.ProcessProposal(ctx, signedProposal, c.grpcOptions...)
	if err != nil {
		return nil, err
	}

	if err = proposal.CheckSuccessfulResponse(proposalResponse); err != nil {
		return nil, err
	}

	result := &lifecycle.QueryInstalledChaincodesResult{}
	if err = proto.Unmarshal(proposalResponse.GetResponse().GetPayload(), result); err != nil {
		return nil, fmt.Errorf("failed to deserialize query installed chaincode result: %w", err)
	}

	return result, nil
}

func (c *command) validate() error {
	if c.grpcClient == nil {
		return errors.New("no gRPC client supplied")
	}

	return nil
}

func (c *command) signedProposal() (*peer.SignedProposal, error) {
	argBytes, err := c.queryInstalledChaincodesArgsBytes()
	if err != nil {
		return nil, err
	}

	proposal, err := proposal.New(
		c.signingID,
		common.LifecycleChaincodeName,
		queryInstalledTransactionName,
		proposal.WithBytesArguments(argBytes),
	)
	if err != nil {
		return nil, err
	}

	proposalBytes, err := proto.Marshal(proposal)
	if err != nil {
		return nil, err
	}

	signature, err := c.signingID.Sign(proposalBytes)
	if err != nil {
		return nil, err
	}

	signedProposal := &peer.SignedProposal{
		ProposalBytes: proposalBytes,
		Signature:     signature,
	}
	return signedProposal, nil
}

func (c *command) queryInstalledChaincodesArgsBytes() ([]byte, error) {
	installArgs := &lifecycle.QueryInstalledChaincodesArgs{}
	return proto.Marshal(installArgs)
}

type Option = func(*command) error

// WithClientConnection uses the supplied gRPC client connection. This should be shared by all commands
// connecting to the same network node.
func WithClientConnection(clientConnection grpc.ClientConnInterface) Option {
	return func(c *command) error {
		c.grpcClient = peer.NewEndorserClient(clientConnection)
		return nil
	}
}

// WithCallOptions specifies the gRPC call options to be used.
func WithCallOptions(options ...grpc.CallOption) Option {
	return func(c *command) error {
		c.grpcOptions = append(c.grpcOptions, options...)
		return nil
	}
}
