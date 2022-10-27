package main

import (
	"os"
	"time"

	"github.com/bestbeforetoday/fabric-admin/pkg/chaincode/install"
	"golang.org/x/net/context"
)

func main() {
	grpcConnection := newGrpcConnection()
	defer grpcConnection.Close()

	id := newIdentity()
	sign := newSign()

	chaincodePackage, err := os.ReadFile("../basic.tar.gz")
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	err = install.Install(
		ctx,
		id,
		install.WithClientConnection(grpcConnection),
		install.WithSign(sign),
		install.WithChaincodePackageBytes(chaincodePackage),
	)
	if err != nil {
		panic(err)
	}
}
