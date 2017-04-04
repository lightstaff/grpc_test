package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/lightstaff/grpc_test/protobuf"

	"google.golang.org/grpc"
)

func main() {
	g := grpc.NewServer()
	s := &Service{}

	pb.RegisterGRPCTestServcieServer(g, s)

	errC := make(chan error)

	go func() {
		lis, err := net.Listen("tcp", ":18080")
		if err != nil {
			errC <- err
		}

		if err := g.Serve(lis); err != nil {
			errC <- err
		}
	}()

	quitC := make(chan os.Signal)
	signal.Notify(quitC, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errC:
		panic(err)
	case <-quitC:
		g.Stop()
	}
}
