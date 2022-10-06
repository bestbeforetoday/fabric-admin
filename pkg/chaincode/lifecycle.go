package chaincode

import (
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"google.golang.org/grpc"
)

type Lifecycle struct {
	grpcClient peer.EndorserClient
}

func NewLifecycle(clientConnection grpc.ClientConnInterface) *Lifecycle {
	return &Lifecycle{
		grpcClient: peer.NewEndorserClient(clientConnection),
	}
}
