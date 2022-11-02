/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package queryinstalled

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer/lifecycle"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

//go:generate mockgen -destination ./endorser_mock_test.go -package ${GOPACKAGE} github.com/hyperledger/fabric-protos-go-apiv2/peer EndorserClient
//go:generate mockgen -destination ./signingidentity_mock_test.go -package ${GOPACKAGE} github.com/bestbeforetoday/fabric-admin/pkg/identity SigningIdentity

func WithEndorserClient(grpcClient peer.EndorserClient) Option {
	return func(b *command) error {
		b.grpcClient = grpcClient
		return nil
	}
}

func NewSigningIdentity(controller *gomock.Controller) *MockSigningIdentity {
	mockIdentity := NewMockSigningIdentity(controller)
	mockIdentity.EXPECT().Creator().AnyTimes()
	mockIdentity.EXPECT().Sign(gomock.Any()).AnyTimes()

	return mockIdentity
}

func NewProposalResponse(status common.Status, message string) *peer.ProposalResponse {
	return &peer.ProposalResponse{
		Response: &peer.Response{
			Status:  int32(status),
			Message: message,
		},
	}
}

func AssertMarshal(t *testing.T, m protoreflect.ProtoMessage) []byte {
	result, err := proto.Marshal(m)
	require.NoError(t, err)
	return result
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

// AssertProtoEqual ensures an expected protobuf message matches an actual message
func AssertProtoEqual(t *testing.T, expected protoreflect.ProtoMessage, actual protoreflect.ProtoMessage) {
	require.True(t, proto.Equal(expected, actual), "Expected %v, got %v", expected, actual)
}

func TestQuery(t *testing.T) {
	t.Run("Missing gRPC connection gives error", func(t *testing.T) {
		controller, ctx := gomock.WithContext(context.Background(), t)
		defer controller.Finish()

		_, err := Query(
			ctx,
			NewSigningIdentity(controller),
		)
		require.ErrorContains(t, err, "gRPC")
	})

	t.Run("Endorser client called with supplied context", func(t *testing.T) {
		controller, ctx := gomock.WithContext(context.Background(), t)
		defer controller.Finish()

		mockEndorser := NewMockEndorserClient(controller)
		mockEndorser.EXPECT().
			ProcessProposal(gomock.Eq(ctx), gomock.Any(), gomock.Any()).
			Return(NewProposalResponse(common.Status_SUCCESS, ""), nil)

		_, err := Query(
			ctx,
			NewSigningIdentity(controller),
			WithEndorserClient(mockEndorser),
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

		_, err := Query(
			ctx,
			NewSigningIdentity(controller),
			WithEndorserClient(mockEndorser),
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

		_, err := Query(
			ctx,
			NewSigningIdentity(controller),
			WithEndorserClient(mockEndorser),
		)

		require.ErrorContainsf(t, err, fmt.Sprintf("%d", expectedStatus), "status code")
		require.ErrorContains(t, err, expectedStatus.String(), "status name")
		require.ErrorContains(t, err, expectedMessage, "message")
	})

	t.Run("Installed chaincodes returned on successful proposal response", func(t *testing.T) {
		expected := &lifecycle.QueryInstalledChaincodesResult{
			InstalledChaincodes: []*lifecycle.QueryInstalledChaincodesResult_InstalledChaincode{
				{
					PackageId: "PACKAGE_ID",
					Label:     "LABEL",
				},
			},
		}
		response := NewProposalResponse(common.Status_SUCCESS, "")
		response.Response.Payload = AssertMarshal(t, expected)

		controller, ctx := gomock.WithContext(context.Background(), t)
		defer controller.Finish()

		mockEndorser := NewMockEndorserClient(controller)
		mockEndorser.EXPECT().
			ProcessProposal(gomock.Eq(ctx), gomock.Any(), gomock.Any()).
			Return(response, nil)

		actual, err := Query(
			ctx,
			NewSigningIdentity(controller),
			WithEndorserClient(mockEndorser),
		)
		require.NoError(t, err)

		AssertProtoEqual(t, expected, actual)
	})

	t.Run("Uses signer", func(t *testing.T) {
		expected := []byte("SIGNATURE")

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

		mockIdentity := NewMockSigningIdentity(controller)
		mockIdentity.EXPECT().Creator().AnyTimes()
		mockIdentity.EXPECT().Sign(gomock.Any()).Return(expected, nil)

		_, err := Query(
			ctx,
			mockIdentity,
			WithEndorserClient(mockEndorser),
		)
		require.NoError(t, err)

		actual := signedProposal.GetSignature()
		require.EqualValues(t, expected, actual)
	})

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

		_, err := Query(
			ctx,
			NewSigningIdentity(controller),
			WithEndorserClient(mockEndorser),
			WithCallOptions(callOption),
		)
		require.NoError(t, err)
	})
}
