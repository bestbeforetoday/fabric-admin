package install

import (
	"context"
	"errors"
	"io"

	"github.com/bestbeforetoday/fabric-admin/internal"
	"github.com/bestbeforetoday/fabric-admin/internal/proposal"
	"github.com/hyperledger/fabric-gateway/pkg/hash"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer/lifecycle"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

const installTransactionName = "InstallChaincode"

func Install(ctx context.Context, id identity.Identity, options ...Option) error {
	installCommand := &command{
		signingID: internal.NewSigningIdentity(id),
	}

	if err := internal.ApplyOptions(installCommand, options...); err != nil {
		return err
	}

	return installCommand.run(ctx)
}

type command struct {
	signingID        *internal.SigningIdentity
	channelName      string
	grpcClient       peer.EndorserClient
	grpcOptions      []grpc.CallOption
	chaincodePackage []byte
}

func (c *command) run(ctx context.Context) error {
	if err := c.validate(); err != nil {
		return err
	}

	signedProposal, err := c.signedProposal()
	if err != nil {
		return err
	}

	proposalResponse, err := c.grpcClient.ProcessProposal(ctx, signedProposal, c.grpcOptions...)
	if err != nil {
		return err
	}

	return internal.CheckSuccessfulProposalResponse(proposalResponse)
}

func (c *command) validate() error {
	if c.grpcClient == nil {
		return errors.New("no gRPC client supplied")
	}
	if c.chaincodePackage == nil {
		return errors.New("no chaincode package supplied")
	}

	return nil
}

func (c *command) signedProposal() (*peer.SignedProposal, error) {
	installArgsBytes, err := c.installChaincodeArgsBytes()
	if err != nil {
		return nil, err
	}

	proposal, err := proposal.New(
		c.signingID,
		c.channelName,
		internal.LifecycleChaincodeName,
		installTransactionName,
		proposal.WithBytesArguments(installArgsBytes),
	)
	if err != nil {
		return nil, err
	}

	signature, proposalBytes, err := c.signingID.SignMessage(proposal)
	if err != nil {
		return nil, err
	}

	signedProposal := &peer.SignedProposal{
		ProposalBytes: proposalBytes,
		Signature:     signature,
	}
	return signedProposal, nil
}

func (c *command) installChaincodeArgsBytes() ([]byte, error) {
	installArgs := &lifecycle.InstallChaincodeArgs{
		ChaincodeInstallPackage: c.chaincodePackage,
	}
	return proto.Marshal(installArgs)
}

type Option = func(*command) error

// WithSign uses the supplied signing implementation to sign messages.
func WithSign(sign identity.Sign) Option {
	return func(c *command) error {
		c.signingID.Sign = sign
		return nil
	}
}

// WithHash uses the supplied hashing implementation to generate digital signatures.
func WithHash(hash hash.Hash) Option {
	return func(c *command) error {
		c.signingID.Hash = hash
		return nil
	}
}

// WithClientConnection uses the supplied gRPC client connection. This should be shared by all commands
// connecting to the same network node.
func WithClientConnection(clientConnection grpc.ClientConnInterface) Option {
	return func(c *command) error {
		c.grpcClient = peer.NewEndorserClient(clientConnection)
		return nil
	}
}

// WithChaincodePackage supplies the chaincode package to be installed.
func WithChaincodePackage(chaincodePackageReader io.Reader) Option {
	return func(c *command) error {
		chaincodePackage, err := io.ReadAll(chaincodePackageReader)
		if err != nil {
			return err
		}

		return WithChaincodePackageBytes(chaincodePackage)(c)
	}
}

// WithChaincodePackageBytes supplies the chaincode package to be installed.
func WithChaincodePackageBytes(chaincodePackage []byte) Option {
	return func(c *command) error {
		c.chaincodePackage = chaincodePackage
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
