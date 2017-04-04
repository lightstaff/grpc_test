package main

import (
	"io"
	"strings"

	pb "github.com/lightstaff/grpc_test/protobuf"

	netCtx "golang.org/x/net/context"
)

// Service model
type Service struct{}

// 単純にHelloと返す
func (s *Service) GetHello(ctx netCtx.Context, e *pb.Empty) (*pb.ReplyModel, error) {
	return &pb.ReplyModel{
		Result: "Hello",
	}, nil
}

// stream経由で受けた文字列を大文字化して返す
func (s *Service) UpperCharacters(stream pb.GRPCTestServcie_UpperCharactersServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if err := stream.Send(&pb.ReplyModel{
			Result: strings.ToUpper(req.Message),
		}); err != nil {
			return err
		}
	}

	return nil
}
