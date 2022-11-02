/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package install

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer/lifecycle"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

//go:generate mockgen -destination ./endorser_mock_test.go -package ${GOPACKAGE} github.com/hyperledger/fabric-protos-go-apiv2/peer EndorserClient
//go:generate mockgen -destination ./identity_mock_test.go -package ${GOPACKAGE} github.com/hyperledger/fabric-gateway/pkg/identity Identity

func WithEndorserClient(grpcClient peer.EndorserClient) Option {
	return func(b *command) error {
		b.grpcClient = grpcClient
		return nil
	}
}

func NewIdentity(controller *gomock.Controller) *MockIdentity {
	mockIdentity := NewMockIdentity(controller)
	mockIdentity.EXPECT().MspID().AnyTimes()
	mockIdentity.EXPECT().Credentials().AnyTimes()

	return mockIdentity
}

func NewSign(result []byte) identity.Sign {
	return func(_ []byte) ([]byte, error) {
		return result, nil
	}
}

func NewProposalResponse(status common.Status, message string) *peer.ProposalResponse {
	return &peer.ProposalResponse{
		Response: &peer.Response{
			Status:  int32(status),
			Message: message,
		},
	}
}

// AssertUnmarshal ensures that a protobuf is umarshaled without error
func AssertUnmarshal(t *testing.T, b []byte, m protoreflect.ProtoMessage) {
	err := proto.Unmarshal(b, m)
	require.NoError(t, err)
}

// AssertUnmarshalProposalPayload ensures that a ChaincodeProposalPayload protobuf is umarshalled without error
func AssertUnmarshalProposalPayload(t *testing.T, signedProposal *peer.SignedProposal) *peer.ChaincodeProposalPayload {
	proposal := &peer.Proposal{}
	AssertUnmarshal(t, signedProposal.ProposalBytes, proposal)

	payload := &peer.ChaincodeProposalPayload{}
	AssertUnmarshal(t, proposal.Payload, payload)

	return payload
}

// AssertUnmarshalInvocationSpec ensures that a ChaincodeInvocationSpec protobuf is umarshalled without error
func AssertUnmarshalInvocationSpec(t *testing.T, signedProposal *peer.SignedProposal) *peer.ChaincodeInvocationSpec {
	proposal := &peer.Proposal{}
	AssertUnmarshal(t, signedProposal.ProposalBytes, proposal)

	payload := &peer.ChaincodeProposalPayload{}
	AssertUnmarshal(t, proposal.Payload, payload)

	input := &peer.ChaincodeInvocationSpec{}
	AssertUnmarshal(t, payload.Input, input)

	return input
}

func TestInstall(t *testing.T) {
	signature := []byte("SIGNATURE")
	sign := NewSign(signature)
	chaincodePackage := []byte("CHAINCODE_PACKAGE")

	t.Run("Missing gRPC connection gives error", func(t *testing.T) {
		controller, ctx := gomock.WithContext(context.Background(), t)
		defer controller.Finish()

		err := Install(
			ctx,
			NewIdentity(controller),
			WithSign(sign),
			WithChaincodePackageBytes(chaincodePackage),
		)
		require.ErrorContains(t, err, "gRPC")
	})

	t.Run("Missing signer gives error", func(t *testing.T) {
		controller, ctx := gomock.WithContext(context.Background(), t)
		defer controller.Finish()

		mockEndorser := NewMockEndorserClient(controller)
		mockEndorser.EXPECT().
			ProcessProposal(gomock.Any(), gomock.Any(), gomock.Any()).
			Times(0)

		err := Install(
			ctx,
			NewIdentity(controller),
			WithEndorserClient(mockEndorser),
			WithChaincodePackageBytes(chaincodePackage),
		)
		require.ErrorContains(t, err, "sign")
	})

	t.Run("Missing chaincode package gives error", func(t *testing.T) {
		controller, ctx := gomock.WithContext(context.Background(), t)
		defer controller.Finish()

		mockEndorser := NewMockEndorserClient(controller)
		mockEndorser.EXPECT().
			ProcessProposal(gomock.Any(), gomock.Any(), gomock.Any()).
			Times(0)

		err := Install(
			ctx,
			NewIdentity(controller),
			WithEndorserClient(mockEndorser),
			WithSign(sign),
		)
		require.ErrorContains(t, err, "chaincode")
	})

	t.Run("Endorser client called with supplied context", func(t *testing.T) {
		controller, ctx := gomock.WithContext(context.Background(), t)
		defer controller.Finish()

		mockEndorser := NewMockEndorserClient(controller)
		mockEndorser.EXPECT().
			ProcessProposal(gomock.Eq(ctx), gomock.Any(), gomock.Any()).
			Return(NewProposalResponse(common.Status_SUCCESS, ""), nil)

		err := Install(
			ctx,
			NewIdentity(controller),
			WithEndorserClient(mockEndorser),
			WithSign(sign),
			WithChaincodePackageBytes(chaincodePackage),
		)
		require.NoError(t, err)
	})

	t.Run("Endorser client errors returned", func(t *testing.T) {
		expectedErr := errors.New("EXPECTED_ERROR")

		controller, ctx := gomock.WithContext(context.Background(), t)
		defer controller.Finish()

		mockEndorser := NewMockEndorserClient(controller)
		mockEndorser.EXPECT().
			ProcessProposal(gomock.Eq(ctx), gomock.Any(), gomock.Any()).
			Return(nil, expectedErr)

		err := Install(
			ctx,
			NewIdentity(controller),
			WithEndorserClient(mockEndorser),
			WithSign(sign),
			WithChaincodePackageBytes(chaincodePackage),
		)
		require.EqualError(t, err, expectedErr.Error())
	})

	t.Run("Unsuccessful proposal response gives error", func(t *testing.T) {
		expectedStatus := common.Status_BAD_REQUEST
		expectedMessage := "EXPECTED_ERROR"

		controller, ctx := gomock.WithContext(context.Background(), t)
		defer controller.Finish()

		mockEndorser := NewMockEndorserClient(controller)
		mockEndorser.EXPECT().
			ProcessProposal(gomock.Eq(ctx), gomock.Any(), gomock.Any()).
			Return(NewProposalResponse(expectedStatus, expectedMessage), nil)

		err := Install(
			ctx,
			NewIdentity(controller),
			WithEndorserClient(mockEndorser),
			WithSign(sign),
			WithChaincodePackageBytes(chaincodePackage),
		)

		require.ErrorContainsf(t, err, fmt.Sprintf("%d", expectedStatus), "status code")
		require.ErrorContains(t, err, expectedStatus.String(), "status name")
		require.ErrorContains(t, err, expectedMessage, "message")
	})

	t.Run("Uses signer", func(t *testing.T) {
		expected := []byte("HASH")

		controller, ctx := gomock.WithContext(context.Background(), t)
		defer controller.Finish()

		var signedProposal *peer.SignedProposal
		mockEndorser := NewMockEndorserClient(controller)
		mockEndorser.EXPECT().
			ProcessProposal(gomock.Eq(ctx), gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *peer.SignedProposal, _ ...grpc.CallOption) {
				signedProposal = in
			}).
			Return(NewProposalResponse(common.Status_SUCCESS, ""), nil).
			Times(1)

		err := Install(
			ctx,
			NewIdentity(controller),
			WithEndorserClient(mockEndorser),
			WithSign(func(digest []byte) ([]byte, error) {
				return digest, nil // Use the hash as the signature
			}),
			WithHash(func(message []byte) []byte {
				return expected
			}),
			WithChaincodePackageBytes(chaincodePackage),
		)
		require.NoError(t, err)

		actual := signedProposal.GetSignature()
		require.EqualValues(t, expected, actual)
	})

	t.Run("Uses hash", func(t *testing.T) {
		controller, ctx := gomock.WithContext(context.Background(), t)
		defer controller.Finish()

		var signedProposal *peer.SignedProposal
		mockEndorser := NewMockEndorserClient(controller)
		mockEndorser.EXPECT().
			ProcessProposal(gomock.Eq(ctx), gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *peer.SignedProposal, _ ...grpc.CallOption) {
				signedProposal = in
			}).
			Return(NewProposalResponse(common.Status_SUCCESS, ""), nil).
			Times(1)

		err := Install(
			ctx,
			NewIdentity(controller),
			WithEndorserClient(mockEndorser),
			WithSign(sign),
			WithChaincodePackageBytes(chaincodePackage),
		)
		require.NoError(t, err)

		actual := signedProposal.GetSignature()
		require.EqualValues(t, signature, actual)
	})

	packageTests := []struct {
		name   string
		option Option
	}{
		{
			name:   "Proposal includes supplied chaincode package bytes",
			option: WithChaincodePackageBytes(chaincodePackage),
		},
		{
			name:   "Proposal includes supplied chaincode package reader",
			option: WithChaincodePackage(bytes.NewReader(chaincodePackage)),
		},
	}
	for _, packageTest := range packageTests {
		t.Run(packageTest.name, func(t *testing.T) {
			controller, ctx := gomock.WithContext(context.Background(), t)
			defer controller.Finish()

			var signedProposal *peer.SignedProposal
			mockEndorser := NewMockEndorserClient(controller)
			mockEndorser.EXPECT().
				ProcessProposal(gomock.Eq(ctx), gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, in *peer.SignedProposal, _ ...grpc.CallOption) {
					signedProposal = in
				}).
				Return(NewProposalResponse(common.Status_SUCCESS, ""), nil).
				Times(1)

			err := Install(
				ctx,
				NewIdentity(controller),
				WithEndorserClient(mockEndorser),
				WithSign(sign),
				packageTest.option,
			)
			require.NoError(t, err)

			invocationSpec := AssertUnmarshalInvocationSpec(t, signedProposal)
			args := invocationSpec.GetChaincodeSpec().GetInput().GetArgs()
			require.Len(t, args, 2, "number of arguments")

			chaincodeArgs := &lifecycle.InstallChaincodeArgs{}
			AssertUnmarshal(t, args[1], chaincodeArgs)

			actual := chaincodeArgs.GetChaincodeInstallPackage()
			require.EqualValues(t, chaincodePackage, actual, "chaincode package")
		})
	}

	t.Run("Endorser client called with supplied gRPC call options", func(t *testing.T) {
		callOption := grpc.WaitForReady(true)

		controller, ctx := gomock.WithContext(context.Background(), t)
		defer controller.Finish()

		mockEndorser := NewMockEndorserClient(controller)
		mockEndorser.EXPECT().
			ProcessProposal(
				gomock.Eq(ctx),
				gomock.Any(),
				gomock.InAnyOrder([]grpc.CallOption{
					callOption,
				}),
			).
			Return(NewProposalResponse(common.Status_SUCCESS, ""), nil)

		err := Install(
			ctx,
			NewIdentity(controller),
			WithEndorserClient(mockEndorser),
			WithSign(sign),
			WithChaincodePackageBytes(chaincodePackage),
			WithCallOptions(callOption),
		)
		require.NoError(t, err)
	})
}
