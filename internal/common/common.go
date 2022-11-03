/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package common

const LifecycleChaincodeName = "_lifecycle"

func ApplyOptions[T any, O ~func(*T) error](target *T, options ...O) error {
	for _, option := range options {
		if err := option(target); err != nil {
			return err
		}
	}

	return nil
}
