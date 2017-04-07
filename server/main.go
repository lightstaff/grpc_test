package main

import (
	"net"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	pb "github.com/lightstaff/grpc_test/protobuf"

	"github.com/Sirupsen/logrus"
	netCtx "golang.org/x/net/context"
	"google.golang.org/grpc"
)

func unaryServerInterceptor(logger *logrus.Logger) grpc.UnaryServerInterceptor {
	return func(ctx netCtx.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var err error
		defer func(begin time.Time) {
			method := path.Base(info.FullMethod)
			took := time.Since(begin)
			fields := logrus.Fields{
				"method": method,
				"took":   took,
			}
			if err != nil {
				fields["error"] = err
				logger.WithFields(fields).Error("Failed")
			} else {
				logger.WithFields(fields).Info("Successed")
			}
		}(time.Now())

		reply, hErr := handler(ctx, req)
		if hErr != nil {
			err = hErr
		}

		return reply, err
	}
}

func streamServerInterceptor(logger *logrus.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		var err error
		defer func(begin time.Time) {
			method := path.Base(info.FullMethod)
			took := time.Since(begin)
			fields := logrus.Fields{
				"method": method,
				"took":   took,
			}
			if err != nil {
				fields["error"] = err
				logger.WithFields(fields).Error("Failed")
			} else {
				logger.WithFields(fields).Info("Successed")
			}
		}(time.Now())

		if hErr := handler(srv, stream); err != nil {
			err = hErr
		}

		return err
	}
}

func main() {
	logger := logrus.New()

	ops := make([]grpc.ServerOption, 0)
	ops = append(ops, grpc.UnaryInterceptor(unaryServerInterceptor(logger)))
	ops = append(ops, grpc.StreamInterceptor(streamServerInterceptor(logger)))

	g := grpc.NewServer(ops...)
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

	logger.Info("start server")

	select {
	case err := <-errC:
		logger.Fatal(err.Error())
	case <-quitC:
		g.Stop()
		logger.Info("stop server")
	}
}
