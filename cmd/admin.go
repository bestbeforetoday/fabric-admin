package main

import (
	"fmt"
	"os"
	"time"

	"github.com/bestbeforetoday/fabric-admin/pkg/chaincode/install"
	"github.com/bestbeforetoday/fabric-admin/pkg/chaincode/queryinstalled"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func main() {
	grpcConnection := newGrpcConnection()
	defer grpcConnection.Close()

	r := &runner{
		grpcConnection: grpcConnection,
		id:             newIdentity(),
		sign:           newSign(),
	}

	r.install()
	r.queryInstalled()
}

type runner struct {
	grpcConnection grpc.ClientConnInterface
	id             identity.Identity
	sign           identity.Sign
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
		r.id,
		install.WithClientConnection(r.grpcConnection),
		install.WithSign(r.sign),
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
		r.id,
		queryinstalled.WithClientConnection(r.grpcConnection),
		queryinstalled.WithSign(r.sign),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(result)
}
