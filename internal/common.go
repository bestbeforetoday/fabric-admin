/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package internal

import (
	"fmt"

	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
)

const LifecycleChaincodeName = "_lifecycle"

func ApplyOptions[T any, O ~func(*T) error](target *T, options ...O) error {
	for _, option := range options {
		if err := option(target); err != nil {
			return err
		}
	}

	return nil
}

func CheckSuccessfulProposalResponse(proposalResponse *peer.ProposalResponse) error {
	response := proposalResponse.GetResponse()
	status := response.GetStatus()

	if status < int32(common.Status_SUCCESS) || status >= int32(common.Status_BAD_REQUEST) {
		return fmt.Errorf("unsuccessful response received with status %d (%s): %s", status, common.Status_name[status], response.GetMessage())
	}

	return nil
}
