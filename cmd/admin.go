/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/bestbeforetoday/fabric-admin/pkg/chaincode/install"
	"github.com/bestbeforetoday/fabric-admin/pkg/chaincode/queryinstalled"
	"github.com/bestbeforetoday/fabric-admin/pkg/identity"
	"github.com/hyperledger/fabric-gateway/pkg/hash"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func main() {
	grpcConnection := newGrpcConnection()
	defer grpcConnection.Close()

	r := &runner{
		grpcConnection: grpcConnection,
		signingID:      identity.NewSigningIdentity(newIdentity(), newSign(), hash.SHA256),
	}

	r.install()
	r.queryInstalled()
}

type runner struct {
	grpcConnection grpc.ClientConnInterface
	signingID      identity.SigningIdentity
}

func (r *runner) install() {
	chaincodePackage, err := os.ReadFile("../basic.tar.gz")
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	err = install.Install(
		ctx,
		r.signingID,
		install.WithClientConnection(r.grpcConnection),
		install.WithChaincodePackageBytes(chaincodePackage),
	)
	if err != nil {
		fmt.Printf("Install failed: %v\n", err)
	}
}

func (r *runner) queryInstalled() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := queryinstalled.Query(
		ctx,
		r.signingID,
		queryinstalled.WithClientConnection(r.grpcConnection),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(result)
}
