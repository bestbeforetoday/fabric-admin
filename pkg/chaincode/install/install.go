package install

import (
	"context"
	"errors"

	common "github.com/bestbeforetoday/fabric-admin/pkg/chaincode/internal"
	"github.com/hyperledger/fabric-gateway/pkg/hash"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"google.golang.org/grpc"
)

func Run(ctx context.Context, id identity.Identity, options ...InstallOption) error {
	command := &installCommand{
		signingID: *common.NewSigningIdentity(id),
	}

	if err := common.ApplyOptions(command, options...); err != nil {
		return err
	}

	if err := command.isValid(); err != nil {
		return err
	}

	// Build and sign proposal
	signedProposal := &peer.SignedProposal{}

	_, err := command.grpcClient.ProcessProposal(ctx, signedProposal, command.grpcOptions...)
	if err != nil {
		return err
	}

	// Check proposal response

	return nil
}

type InstallOption = func(*installCommand) error

// WithSign uses the supplied signing implementation to sign messages.
func WithSign(sign identity.Sign) InstallOption {
	return func(command *installCommand) error {
		command.signingID.Sign = sign
		return nil
	}
}

// WithHash uses the supplied hashing implementation to generate digital signatures.
func WithHash(hash hash.Hash) InstallOption {
	return func(command *installCommand) error {
		command.signingID.Hash = hash
		return nil
	}
}

// WithClientConnection uses the supplied gRPC client connection. This should be shared by all commands
// connecting to the same network node.
func WithClientConnection(clientConnection grpc.ClientConnInterface) InstallOption {
	return func(command *installCommand) error {
		command.grpcClient = peer.NewEndorserClient(clientConnection)
		return nil
	}
}

type installCommand struct {
	signingID   common.SigningIdentity
	grpcClient  peer.EndorserClient
	grpcOptions []grpc.CallOption
}

func (command *installCommand) isValid() error {
	if command.grpcClient == nil {
		return errors.New("no gRPC client supplied")
	}

	return nil
}
