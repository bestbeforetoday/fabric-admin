package install

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/stretchr/testify/require"
)

//go:generate mockgen -destination ./endorser_mock_test.go -package ${GOPACKAGE} github.com/hyperledger/fabric-protos-go-apiv2/peer EndorserClient
//go:generate mockgen -destination ./identity_mock_test.go -package ${GOPACKAGE} github.com/hyperledger/fabric-gateway/pkg/identity Identity

// WithClientConnection uses the supplied gRPC client connection. This should be shared by all commands
// connecting to the same network node.
func WithEndorserClient(grpcClient peer.EndorserClient) InstallOption {
	return func(command *installCommand) error {
		command.grpcClient = grpcClient
		return nil
	}
}

func TestInstall(t *testing.T) {
	t.Run("Missing gRPC connection gives error", func(t *testing.T) {
		controller, ctx := gomock.WithContext(context.Background(), t)
		defer controller.Finish()

		err := Run(ctx, NewMockIdentity(controller))
		require.ErrorContains(t, err, "gRPC")
	})

	t.Run("Endorser client called with supplied context", func(t *testing.T) {
		controller, ctx := gomock.WithContext(context.Background(), t)
		defer controller.Finish()

		mockEndorser := NewMockEndorserClient(controller)
		mockEndorser.EXPECT().
			ProcessProposal(gomock.Eq(ctx), gomock.Any(), gomock.Len(0))

		err := Run(ctx, NewMockIdentity(controller), WithEndorserClient(mockEndorser))
		require.NoError(t, err)
	})

	t.Run("Endorser client errors returned", func(t *testing.T) {
		controller, ctx := gomock.WithContext(context.Background(), t)
		defer controller.Finish()

		expectedErr := errors.New("EXPECTED_ERROR")

		mockEndorser := NewMockEndorserClient(controller)
		mockEndorser.EXPECT().
			ProcessProposal(gomock.Eq(ctx), gomock.Any(), gomock.Len(0)).
			Return(nil, expectedErr)

		actualErr := Run(ctx, NewMockIdentity(controller), WithEndorserClient(mockEndorser))
		require.EqualError(t, actualErr, expectedErr.Error())
	})
}
