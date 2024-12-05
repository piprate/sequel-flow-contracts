package iinft

import (
	"fmt"

	"github.com/onflow/flow-go-sdk/access"
	grpcAccess "github.com/onflow/flow-go-sdk/access/grpc"
	"github.com/onflow/flowkit/v2"
	"google.golang.org/grpc"
)

// maxGRPCMessageSize 60mb
const maxGRPCMessageSize = 1024 * 1024 * 60

func NewGrpcClient(baseLoader flowkit.ReaderWriter, network string, opts ...grpcAccess.ClientOption) (access.Client, error) {
	state, err := flowkit.Load([]string{"flow.json"}, baseLoader)
	if err != nil {
		return nil, err
	}

	networkDef, err := state.Networks().ByName(network)
	if err != nil {
		return nil, err
	}

	options := append(
		[]grpcAccess.ClientOption{
			grpcAccess.WithGRPCDialOptions(
				grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxGRPCMessageSize)),
			),
		},
		opts...,
	)

	gClient, err := grpcAccess.NewClient(
		networkDef.Host,
		options...,
	)

	if err != nil || gClient == nil {
		return nil, fmt.Errorf("failed to connect to host %s", networkDef.Host)
	}

	return gClient, nil
}
