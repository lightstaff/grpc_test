package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "github.com/lightstaff/grpc_test/protobuf"
	"google.golang.org/grpc"

	"net/http"

	"github.com/labstack/echo"
)

type ServiceContext struct {
	echo.Context
	ServiceClient pb.GRPCTestServcieClient
}

func serviceContextMiddleware(grpcAddr string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc, err := grpc.Dial(grpcAddr, grpc.WithBlock(), grpc.WithInsecure())
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
			}
			defer cc.Close()

			sc := &ServiceContext{
				Context:       c,
				ServiceClient: pb.NewGRPCTestServcieClient(cc),
			}

			return next(sc)
		}
	}
}

func main() {
	e := echo.New()

	e.Use(serviceContextMiddleware("localhost:18080"))

	e.GET("/hello", GetHello)
	e.POST("/upper-characters", UpperCharacters)

	errC := make(chan error)
	go func() {
		if err := e.Start(":8080"); err != nil {
			errC <- err
		}
	}()

	quitC := make(chan os.Signal)
	signal.Notify(quitC, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errC:
		panic(err)
	case <-quitC:
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := e.Shutdown(shutdownCtx); err != nil {
			errC <- err
		}
	}
}
